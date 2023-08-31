/**
 * Copyright 2023 ByteDance Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package loader

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unsafe"
)

const (
	_MinLC   uint8 = 1
	_PtrSize uint8 = 8
)

const (
	_N_FUNCDATA              = 8
	_INVALID_FUNCDATA_OFFSET = ^uint32(0)
	_FUNC_SIZE               = unsafe.Sizeof(_func{})

	_MINFUNC        = 16 // minimum size for a function
	_BUCKETSIZE     = 256 * _MINFUNC
	_SUBBUCKETS     = 16
	_SUB_BUCKETSIZE = _BUCKETSIZE / _SUBBUCKETS
)

// PCDATA and FUNCDATA table indexes.
//
// See funcdata.h and $GROOT/src/cmd/internal/objabi/funcdata.go.
const (
	_FUNCDATA_ArgsPointerMaps    = 0
	_FUNCDATA_LocalsPointerMaps  = 1
	_FUNCDATA_StackObjects       = 2
	_FUNCDATA_InlTree            = 3
	_FUNCDATA_OpenCodedDeferInfo = 4
	_FUNCDATA_ArgInfo            = 5
	_FUNCDATA_ArgLiveInfo        = 6
	_FUNCDATA_WrapInfo           = 7

	// ArgsSizeUnknown is set in Func.argsize to mark all functions
	// whose argument size is unknown (C vararg functions, and
	// assembly code without an explicit specification).
	// This value is generated by the compiler, assembler, or linker.
	ArgsSizeUnknown = -0x80000000
)

// moduledata used to cache the funcdata and findfuncbucket of one module
var moduleCache = struct {
	m map[*moduledata][]byte
	sync.Mutex
}{
	m: make(map[*moduledata][]byte),
}

// Func contains information about a function.
type Func struct {
	ID          uint8  // see runtime/symtab.go
	Flag        uint8  // see runtime/symtab.go
	ArgsSize    int32  // args byte size
	EntryOff    uint32 // start pc, offset to moduledata.text
	TextSize    uint32 // size of func text
	DeferReturn uint32 // offset of start of a deferreturn call instruction from entry, if any.
	FileIndex   uint32 // index into filetab
	Name        string // name of function

	// PC data
	Pcsp            *Pcdata // PC -> SP delta
	Pcfile          *Pcdata // PC -> file index
	Pcline          *Pcdata // PC -> line number
	PcUnsafePoint   *Pcdata // PC -> unsafe point, must be PCDATA_UnsafePointSafe or PCDATA_UnsafePointUnsafe
	PcStackMapIndex *Pcdata // PC -> stack map index, relative to ArgsPointerMaps and LocalsPointerMaps
	PcInlTreeIndex  *Pcdata // PC -> inlining tree index, relative to InlTree
	PcArgLiveIndex  *Pcdata // PC -> arg live index, relative to ArgLiveInfo

	// Func data, must implement encoding.BinaryMarshaler
	ArgsPointerMaps    encoding.BinaryMarshaler // concrete type: *StackMap
	LocalsPointerMaps  encoding.BinaryMarshaler // concrete type: *StackMap
	StackObjects       encoding.BinaryMarshaler
	InlTree            encoding.BinaryMarshaler
	OpenCodedDeferInfo encoding.BinaryMarshaler
	ArgInfo            encoding.BinaryMarshaler
	ArgLiveInfo        encoding.BinaryMarshaler
	WrapInfo           encoding.BinaryMarshaler
}

func getOffsetOf(data interface{}, field string) uintptr {
	t := reflect.TypeOf(data)
	fv, ok := t.FieldByName(field)
	if !ok {
		panic(fmt.Sprintf("field %s not found in struct %s", field, t.Name()))
	}
	return fv.Offset
}

func rnd(v int64, r int64) int64 {
	if r <= 0 {
		return v
	}
	v += r - 1
	c := v % r
	if c < 0 {
		c += r
	}
	v -= c
	return v
}

var (
	byteOrder binary.ByteOrder = binary.LittleEndian
)

func funcNameParts(name string) (string, string, string) {
	i := strings.IndexByte(name, '[')
	if i < 0 {
		return name, "", ""
	}
	// TODO: use LastIndexByte once the bootstrap compiler is >= Go 1.5.
	j := len(name) - 1
	for j > i && name[j] != ']' {
		j--
	}
	if j <= i {
		return name, "", ""
	}
	return name[:i], "[...]", name[j+1:]
}
