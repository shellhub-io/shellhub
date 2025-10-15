package vnc

import (
	"fmt"
	"image"
	"image/color"
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil"
)

// X11Display implements the Display interface for X11/Xorg
type X11Display struct {
	conn    *xgb.Conn
	xu      *xgbutil.XUtil
	screen  *xproto.ScreenInfo
	rootWin xproto.Window
	width   uint16
	height  uint16
}

// NewX11Display creates a new X11 display handler
func NewX11Display(displayName string) (*X11Display, error) {
	// Connect to X server
	conn, err := xgb.NewConnDisplay(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %v", err)
	}

	// Create XUtil connection for higher-level operations
	xu, err := xgbutil.NewConnDisplay(displayName)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create XUtil connection: %v", err)
	}

	// Get default screen
	setup := xproto.Setup(conn)
	if len(setup.Roots) == 0 {
		conn.Close()
		xu.Conn().Close()
		return nil, fmt.Errorf("no screens available")
	}

	screen := &setup.Roots[0]

	display := &X11Display{
		conn:    conn,
		xu:      xu,
		screen:  screen,
		rootWin: screen.Root,
		width:   screen.WidthInPixels,
		height:  screen.HeightInPixels,
	}

	return display, nil
}

// Close closes the X11 connection
func (d *X11Display) Close() error {
	if d.xu != nil {
		d.xu.Conn().Close()
	}
	if d.conn != nil {
		d.conn.Close()
	}
	return nil
}

// GetScreenSize returns the screen dimensions
func (d *X11Display) GetScreenSize() (width, height uint16) {
	return d.width, d.height
}

// CaptureScreen captures the current screen content
func (d *X11Display) CaptureScreen() ([]uint8, error) {
	// Get image of the root window
	reply, err := xproto.GetImage(d.conn, xproto.ImageFormatZPixmap,
		xproto.Drawable(d.rootWin), 0, 0, d.width, d.height, 0xffffffff).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get screen image: %v", err)
	}

	// Convert X11 image data to VNC format
	imageData := reply.Data
	vncData := make([]uint8, len(imageData))

	// X11 typically uses BGRA format, VNC expects RGBA
	// Convert pixel format if necessary
	for i := 0; i < len(imageData); i += 4 {
		if i+3 < len(imageData) {
			// X11: BGRA -> VNC: RGBA
			vncData[i] = imageData[i+2]   // R
			vncData[i+1] = imageData[i+1] // G
			vncData[i+2] = imageData[i]   // B
			vncData[i+3] = imageData[i+3] // A
		}
	}

	return vncData, nil
}

// SendKeyEvent sends a keyboard event to X11
func (d *X11Display) SendKeyEvent(key uint32, down bool) error {
	// Convert VNC keysym to X11 keycode
	keycode := d.keysymToKeycode(key)
	if keycode == 0 {
		// Key not found, ignore
		return nil
	}

	// For now, just log the key event (requires X11 extensions for proper implementation)
	log.Printf("Key event: key=%d, down=%t, keycode=%d", key, down, keycode)
	return nil
}

// SendPointerEvent sends a mouse event to X11
func (d *X11Display) SendPointerEvent(x, y uint16, buttonMask uint8) error {
	// For now, just log the mouse event (requires X11 extensions for proper implementation)
	log.Printf("Pointer event: x=%d, y=%d, buttons=%d", x, y, buttonMask)
	return nil
}

// keysymToKeycode converts VNC keysym to X11 keycode
func (d *X11Display) keysymToKeycode(keysym uint32) uint8 {
	// This is a simplified mapping. A complete implementation would use
	// the X11 keymap to do proper conversion.

	// Common ASCII keys
	if keysym >= 0x20 && keysym <= 0x7E {
		// Try to find the keycode for this keysym
		minKeycode := d.setup().MinKeycode
		maxKeycode := d.setup().MaxKeycode

		for keycode := minKeycode; keycode <= maxKeycode; keycode++ {
			// This is a simplified approach - in practice you'd use
			// XGetKeyboardMapping or similar to get the proper mapping
			if uint32(keycode) == keysym-0x20+uint32(minKeycode) {
				return uint8(keycode)
			}
		}
	}

	// Special keys mapping
	specialKeys := map[uint32]uint8{
		0xFF08: 22,  // BackSpace
		0xFF09: 23,  // Tab
		0xFF0D: 36,  // Return
		0xFF1B: 9,   // Escape
		0xFF20: 64,  // Space (though it should be handled above)
		0xFFE1: 50,  // Shift_L
		0xFFE2: 62,  // Shift_R
		0xFFE3: 37,  // Control_L
		0xFFE4: 105, // Control_R
		0xFFE9: 64,  // Alt_L
		0xFFEA: 108, // Alt_R
		0xFF51: 113, // Left arrow
		0xFF52: 111, // Up arrow
		0xFF53: 114, // Right arrow
		0xFF54: 116, // Down arrow
	}

	if keycode, exists := specialKeys[keysym]; exists {
		return keycode
	}

	return 0 // Unknown key
}

// setup returns the X11 setup information
func (d *X11Display) setup() *xproto.SetupInfo {
	return xproto.Setup(d.conn)
}

// MockDisplay is a simple mock implementation for testing
type MockDisplay struct {
	width  uint16
	height uint16
	pixels []uint8
}

// NewMockDisplay creates a mock display for testing
func NewMockDisplay(width, height uint16) *MockDisplay {
	pixelCount := int(width) * int(height) * 4 // RGBA
	pixels := make([]uint8, pixelCount)

	// Fill with a simple pattern
	for i := 0; i < pixelCount; i += 4 {
		pixels[i] = 100   // R
		pixels[i+1] = 150 // G
		pixels[i+2] = 200 // B
		pixels[i+3] = 255 // A
	}

	return &MockDisplay{
		width:  width,
		height: height,
		pixels: pixels,
	}
}

// GetScreenSize returns the mock screen size
func (m *MockDisplay) GetScreenSize() (width, height uint16) {
	return m.width, m.height
}

// CaptureScreen returns the mock screen content
func (m *MockDisplay) CaptureScreen() ([]uint8, error) {
	// Return a copy of the pixel data
	data := make([]uint8, len(m.pixels))
	copy(data, m.pixels)
	return data, nil
}

// SendKeyEvent handles key events (mock implementation)
func (m *MockDisplay) SendKeyEvent(key uint32, down bool) error {
	log.Printf("Mock key event: key=%d, down=%t", key, down)
	return nil
}

// SendPointerEvent handles mouse events (mock implementation)
func (m *MockDisplay) SendPointerEvent(x, y uint16, buttonMask uint8) error {
	log.Printf("Mock pointer event: x=%d, y=%d, buttons=%d", x, y, buttonMask)
	return nil
}

// CreateTestPattern creates a test pattern image
func CreateTestPattern(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a gradient pattern
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))

			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}

	return img
}
