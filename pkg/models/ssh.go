package models

// SSHCommand is the payload of an "exec" request: the single command line the client asked to run
// instead of an interactive shell.
type SSHCommand struct {
	Command string `json:"command"`
}

// SSHSubsystem is the payload of a "subsystem" request, which is how SFTP and friends are asked
// for by name rather than by command line.
type SSHSubsystem struct {
	Subsystem string `json:"subsystem"`
}

// SSHExitStatus is the payload of an "exit-status" request: the process's exit code, sent when it
// ended normally rather than on a signal.
type SSHExitStatus struct {
	Status uint32 `json:"status"`
}

// SSHSignal is the payload of an "exit-signal" request, sent instead of an exit status when the
// process was killed. Dumped says whether a core file was written.
type SSHSignal struct {
	Name    uint32 `json:"status"`
	Dumped  bool   `json:"dumped"`
	Message string `json:"message"`
	Lang    string `json:"lang"`
}

// SSHWindowChange is the payload of a "window-change" request, sent every time the client's
// terminal is resized. Columns and rows are what the application sees; width and height are pixels
// and are usually zero.
type SSHWindowChange struct {
	Columns uint32 `json:"columns"`
	Rows    uint32 `json:"rows"`
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
}

// SSHPty is the payload of a "pty-req" request, the terminal the client asks for before a shell.
// It repeats SSHWindowChange's fields rather than embedding it because ssh.Unmarshal maps a flat
// wire format onto the struct and cannot descend into a nested one.
type SSHPty struct {
	Term    string `json:"term"`
	Columns uint32 `json:"columns"`
	Rows    uint32 `json:"rows"`
	Width   uint32 `json:"width"`
	Height  uint32 `json:"height"`
	// Not persisted (json:"-"); kept only so gossh.Unmarshal can consume it.
	Modelist []byte `json:"-"`
}

// SSHPtyOutput carries the bytes a terminal produced. It has no SSH request of its own — the
// protocol sends this as channel data — and exists so recorded output can be stored as an event
// alongside the requests.
type SSHPtyOutput struct {
	Output string `json:"output"`
}
