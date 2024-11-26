package metrics

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

type LscpuEntry struct {
	Field    string       `json:"field"`
	Data     string       `json:"data"`
	Children []LscpuEntry `json:"children,omitempty"`
}

type Lscpu struct {
	Lscpu []LscpuEntry `json:"lscpu"`
}

func parseJSON(r io.Reader, v interface{}) (interface{}, error) {
	// Read the entire content of the io.Reader first to check for errors even if valid json is first
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Errorf("error reading from io.Reader: %v", err)
	}

	err = json.Unmarshal(buf, v)
	if err != nil {
		return nil, errors.Errorf("couldn't parse JSON: %v", err)
	}
	return v, nil
}
