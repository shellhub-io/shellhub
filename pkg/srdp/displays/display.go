package displays

// Display interface for screen capture
type Display interface {
	// TODO: Evaluate context usage in Display methods.
	GetScreenSize() (width, height uint16)
	CaptureScreen() ([]uint8, error)
	SendKeyEvent(key uint32, down bool) error
	SendPointerEvent(x, y uint16, buttonMask uint8) error
}
