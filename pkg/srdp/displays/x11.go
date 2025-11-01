package displays

import (
	"fmt"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"
	"github.com/jezek/xgbutil"
)

type X11Display struct {
	conn   *xgb.Conn
	xu     *xgbutil.XUtil
	setup  *xproto.SetupInfo
	screen *xproto.ScreenInfo
	window xproto.Window
	width  uint16
	height uint16

	lastButtonMask uint8 // Track previous button state for press/release detection
}

var _ Display = (*X11Display)(nil)

func NewX11Display(displayName string) (*X11Display, error) {
	conn, err := xgb.NewConnDisplay(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to X server: %v", err)
	}

	xu, err := xgbutil.NewConnDisplay(displayName)
	if err != nil {
		conn.Close()

		return nil, fmt.Errorf("failed to create XUtil connection: %v", err)
	}

	setup := xproto.Setup(conn)
	if len(setup.Roots) == 0 {
		conn.Close()
		xu.Conn().Close()

		return nil, fmt.Errorf("no screens available")
	}

	screen := &setup.Roots[0]

	if err := xtest.Init(conn); err != nil {
		conn.Close()
		xu.Conn().Close()

		return nil, fmt.Errorf("failed to initialize XTest extension: %v", err)
	}

	display := &X11Display{
		conn:   conn,
		xu:     xu,
		setup:  setup,
		screen: screen,
		window: screen.Root,
		width:  screen.WidthInPixels,
		height: screen.HeightInPixels,
	}

	return display, nil
}

func (d *X11Display) Close() error {
	if d.xu != nil {
		d.xu.Conn().Close()
	}

	if d.conn != nil {
		d.conn.Close()
	}

	return nil
}

func (d *X11Display) GetScreenSize() (width, height uint16) {
	return d.width, d.height
}

func (d *X11Display) CaptureScreen() ([]uint8, error) {
	reply, err := xproto.GetImage(d.conn, xproto.ImageFormatZPixmap,
		xproto.Drawable(d.window), 0, 0, d.width, d.height, 0xffffffff).Reply()
	if err != nil {
		return nil, fmt.Errorf("failed to get screen image: %v", err)
	}

	return reply.Data, nil
}

// SendKeyEvent sends a keyboard event to X11 using xtest
func (d *X11Display) SendKeyEvent(key uint32, down bool) error {
	keycode := d.keysymToKeycode(key)
	if keycode == 0 {
		return nil
	}

	err := xtest.FakeInputChecked(d.conn, xproto.KeyPress, keycode, 0, d.window, 0, 0, 0).Check()
	if err != nil {
		return fmt.Errorf("failed to send key press event: %v", err)
	}

	if !down {
		err = xtest.FakeInputChecked(d.conn, xproto.KeyRelease, keycode, 0, d.window, 0, 0, 0).Check()
		if err != nil {
			return fmt.Errorf("failed to send key release event: %v", err)
		}
	}

	return nil
}

func (d *X11Display) SendPointerEvent(ux, uy uint16, buttonMask uint8) error {
	x, y := int16(ux), int16(uy) // nolint:gosec

	if err := xtest.FakeInputChecked(d.conn, xproto.MotionNotify, 0, 0, d.window, x, y, 0).Check(); err != nil {
		return fmt.Errorf("failed to move pointer: %v", err)
	}

	// Detect button state changes by comparing with previous state
	buttonChanges := d.lastButtonMask ^ buttonMask

	// Button 1 (left click) - bit 0
	if buttonChanges&0x01 != 0 {
		if buttonMask&0x01 != 0 {
			// Button pressed
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonPress, 1, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send left button press: %v", err)
			}
		} else {
			// Button released
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonRelease, 1, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send left button release: %v", err)
			}
		}
	}

	// Button 2 (middle click) - bit 1
	if buttonChanges&0x02 != 0 {
		if buttonMask&0x02 != 0 {
			// Button pressed
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonPress, 2, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send middle button press: %v", err)
			}
		} else {
			// Button released
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonRelease, 2, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send middle button release: %v", err)
			}
		}
	}

	// Button 3 (right click) - bit 2
	if buttonChanges&0x04 != 0 {
		if buttonMask&0x04 != 0 {
			// Button pressed
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonPress, 3, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send right button press: %v", err)
			}
		} else {
			// Button released
			if err := xtest.FakeInputChecked(d.conn, xproto.ButtonRelease, 3, 0, d.window, x, y, 0).Check(); err != nil {
				return fmt.Errorf("failed to send right button release: %v", err)
			}
		}
	}

	// Update the last button state
	d.lastButtonMask = buttonMask

	return nil
}

func (d *X11Display) keysymToKeycode(keysym uint32) uint8 {
	minKeycode := d.setup.MinKeycode
	maxKeycode := d.setup.MaxKeycode

	count := byte(maxKeycode - minKeycode + 1)
	reply, err := xproto.GetKeyboardMapping(d.conn, minKeycode, count).Reply()
	if err != nil {
		// TODO: Add debug logging here.
		return 0
	}

	// NOTE: Search through the keyboard mapping for the matching keysym.
	if reply != nil && len(reply.Keysyms) > 0 {
		keysymsPerKeycode := int(reply.KeysymsPerKeycode)

		for keycode := minKeycode; keycode <= maxKeycode; keycode++ {
			// FIXME: Weird infinite loop workaround.
			if keycode == maxKeycode {
				fmt.Printf("Reached max keycode %d, stopping search\n", maxKeycode)

				break
			}

			baseIdx := int(keycode-minKeycode) * keysymsPerKeycode
			for i := 0; i < keysymsPerKeycode && baseIdx+i < len(reply.Keysyms); i++ {
				if uint32(reply.Keysyms[baseIdx+i]) == keysym {
					return uint8(keycode)
				}
			}
		}
	}

	return 0 // Unknown key
}
