package common

// COMStruct is the serial configuration.
type COMStruct struct {
	SerialPort string `json:"serialPort"`
	BaudRate   int64  `json:"baudRate"`
	DataBits   int64  `json:"dataBits"`
	Parity     string `json:"parity"`
	StopBits   int64  `json:"stopBits"`
}

type DownloadRequest struct {
	Path    string `json:"path"`
	Segment string `json:"segment"`
}

type ExecuteRequest struct {
	ID      uint64 `json:"id"`
	Segment string `json:"segment"`
	Start   int    `json:"strat"`
	End     int    `json:"end"`
	Period  int    `json:"period"`
}
