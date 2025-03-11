package models

type SSHCommand struct {
	Command string `json:"command"`
}

type SSHSubsystem struct {
	Subsystem string `json:"subsystem"`
}

type SSHExitStatus struct {
	Status uint32 `json:"status"`
}

type SSHSignal struct {
	Name    uint32 `json:"status"`
	Dumped  bool   `json:"dumped"`
	Message string `json:"message"`
	Lang    string `json:"lang"`
}

type SSHWindowChange struct {
	Columns uint32 `json:"columns"`
	Rows    uint32 `json:"rows"`
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
}

// NOTE: [SSHPty] cannot use [SSHWindowChange] inside itself due [ssh.Unmarshal] issues.
type SSHPty struct {
	Term     string `json:"term"`
	Columns  uint32 `json:"columns" `
	Rows     uint32 `json:"rows"`
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	Modelist []byte `json:"modelist"`
}

type SSHPtyOutput struct {
	Output string `json:"output"`
}
