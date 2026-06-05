package share

import "encoding/json"

// control is the envelope for non-output (text) frames exchanged over the share websocket.
// Output (PTY bytes) travels as raw binary frames; everything else is a JSON control frame.
type control struct {
	Kind     string `json:"kind"`
	Cols     int    `json:"cols,omitempty"`
	Rows     int    `json:"rows,omitempty"`
	Writable bool   `json:"writable,omitempty"`
	Name     string `json:"name,omitempty"`
}

const (
	controlKindResize = "resize"
	controlKindInit   = "init"
)

// encodeResize encodes a terminal size change as a JSON control frame.
func encodeResize(dim Dimensions) ([]byte, error) {
	return json.Marshal(control{Kind: controlKindResize, Cols: dim.Cols, Rows: dim.Rows})
}

// encodeInit encodes the handshake frame sent to a guest on connect, carrying the share's label
// and whether it is collaborative (writable).
func encodeInit(name string, writable bool) ([]byte, error) {
	return json.Marshal(control{Kind: controlKindInit, Writable: writable, Name: name})
}

// decodeControl parses a JSON control frame received from the producer (agent).
func decodeControl(data []byte) (control, error) {
	var ctrl control
	err := json.Unmarshal(data, &ctrl)

	return ctrl, err
}
