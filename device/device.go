/*
Copyright 2020 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package device

import (
	"sync"
	"time"

	"github.com/smilelinkd/bowexecutor/driver"
	"github.com/smilelinkd/bowexecutor/pkg/common"
)

var wg sync.WaitGroup

// InitBow initialize bow client
func InitBow() (client *driver.DigitalbowClient, err error) {
	COM := common.COMStruct{
		SerialPort: "/dev/ttyS0",
		BaudRate:   115200,
		DataBits:   8,
		Parity:     "even",
		StopBits:   1,
	}
	RTUConfig := driver.BowRTUConfig{
		SerialName: COM.SerialPort,
		BaudRate:   int(COM.BaudRate),
		DataBits:   int(COM.DataBits),
		StopBits:   int(COM.StopBits),
		Parity:     COM.Parity,
		Timeout:    5 * time.Second,
	}
	client, _ = driver.NewClient(RTUConfig)
	client.Client.Init()
	return client, nil
}
