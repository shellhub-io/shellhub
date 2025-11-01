package encoders

/*
#cgo pkg-config: x264
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <x264.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

type H264 struct {
	handle   *C.x264_t
	params   C.x264_param_t
	picIn    C.x264_picture_t
	picOut   C.x264_picture_t
	frameNum int64 // Track frame count for PTS calculation
}

var _ Encoder = (*H264)(nil)

type H264Params struct {
	// https://trac.ffmpeg.org/wiki/Encode/H.264#Preset
	Preset string
	// https://trac.ffmpeg.org/wiki/Encode/H.264#Tune
	Tune string
	// https://trac.ffmpeg.org/wiki/Encode/H.264#Profile
	Profile string
	Bitrate uint
}

var DefaultH264Params = H264Params{
	Preset:  "veryfast",
	Tune:    "zerolatency",
	Profile: "baseline",
	Bitrate: 1000,
}

type H264Option func(*H264Params)

func NewH264(width, height int, fps uint32, ops ...H264Option) (*H264, error) {
	var enc H264

	params := DefaultH264Params
	for _, op := range ops {
		op(&params)
	}

	preset := C.CString(params.Preset)
	tune := C.CString(params.Tune)
	defer C.free(unsafe.Pointer(preset))
	defer C.free(unsafe.Pointer(tune))

	if C.x264_param_default_preset(&enc.params, preset, tune) != 0 {
		return nil, errors.New("failed to apply preset")
	}

	enc.params.i_csp = C.X264_CSP_I420
	enc.params.i_width = C.int(width)
	enc.params.i_height = C.int(height)
	enc.params.b_repeat_headers = 1
	enc.params.b_annexb = 1

	enc.params.i_fps_num = C.uint(fps)
	enc.params.i_fps_den = 1

	enc.params.i_timebase_den = C.uint(fps) // Matches fps: each PTS increment = 1 frame
	enc.params.i_timebase_num = 1

	// TODO: Adjust bitrate settings as needed on runtime if the sharing is too slow.
	enc.params.rc.i_rc_method = C.X264_RC_ABR
	enc.params.rc.i_bitrate = C.int(params.Bitrate)
	enc.params.rc.i_vbv_buffer_size = C.int(params.Bitrate * 2)
	enc.params.rc.i_vbv_max_bitrate = C.int(params.Bitrate * 2)

	profile := C.CString(params.Profile)
	defer C.free(unsafe.Pointer(profile))

	if C.x264_param_apply_profile(&enc.params, profile) != 0 {
		return nil, errors.New("failed to apply profile")
	}

	if C.x264_picture_alloc(&enc.picIn, enc.params.i_csp, enc.params.i_width, enc.params.i_height) != 0 {
		return nil, errors.New("failed to allocate picture")
	}

	enc.handle = C.x264_encoder_open(&enc.params)
	if enc.handle == nil {
		C.x264_picture_clean(&enc.picIn)
		return nil, errors.New("failed to open encoder")
	}

	enc.frameNum = 0

	return &enc, nil
}

const EncoderH264 = 1

func (e *H264) Code() uint8 {
	return EncoderH264
}

// Encode encodes one RGBA frame to H.264
func (e *H264) Encode(width, height uint16, data []byte) ([]byte, error) {
	// TODO: Optimize!
	y, u, v := rgbToYUV420(int(width), int(height), data)

	C.memcpy(unsafe.Pointer(e.picIn.img.plane[0]), unsafe.Pointer(&y[0]), C.size_t(len(y)))
	C.memcpy(unsafe.Pointer(e.picIn.img.plane[1]), unsafe.Pointer(&u[0]), C.size_t(len(u)))
	C.memcpy(unsafe.Pointer(e.picIn.img.plane[2]), unsafe.Pointer(&v[0]), C.size_t(len(v)))

	// FIXME: Set presentation timestamp.
	e.picIn.i_pts = C.int64_t(e.frameNum)
	e.frameNum++

	var nal *C.x264_nal_t
	var iNal C.int
	size := C.x264_encoder_encode(e.handle, &nal, &iNal, &e.picIn, &e.picOut)
	if size < 0 {
		return nil, errors.New("encode failed")
	}
	if size == 0 {
		return nil, nil // Delayed frame
	}

	return C.GoBytes(unsafe.Pointer(nal.p_payload), size), nil
}

// Flush returns delayed frames
func (e *H264) Flush() ([][]byte, error) {
	var frames [][]byte
	for {
		var nal *C.x264_nal_t
		var iNal C.int
		size := C.x264_encoder_encode(e.handle, &nal, &iNal, nil, &e.picOut)
		if size < 0 {
			return nil, errors.New("flush failed")
		}
		if size == 0 {
			break
		}
		frames = append(frames, C.GoBytes(unsafe.Pointer(nal.p_payload), size))
	}
	return frames, nil
}

func (e *H264) Close() {
	if e.handle != nil {
		C.x264_encoder_close(e.handle)
		C.x264_picture_clean(&e.picIn)
		e.handle = nil
	}
}

func rgbToYUV420(width, height int, data []byte) (y, u, v []byte) {
	y = make([]byte, width*height)
	u = make([]byte, (width/2)*(height/2))
	v = make([]byte, (width/2)*(height/2))

	var r, g, b uint16
	var r8, g8, b8 float64
	var yy float64

	for j := range height {
		for i := range width {
			r = uint16(data[(j*width+i)*4+0])<<8 | uint16(data[(j*width+i)*4+0])
			g = uint16(data[(j*width+i)*4+1])<<8 | uint16(data[(j*width+i)*4+1])
			b = uint16(data[(j*width+i)*4+2])<<8 | uint16(data[(j*width+i)*4+2])
			// _ = uint16(data[(j*width+i)*4+3])<<8 | uint16(data[(j*width+i)*4+3]) // Alpha channel, ignored.

			// Convert 16-bit to 8-bit (more precise)
			r8 = float64(r / 257)
			g8 = float64(g / 257)
			b8 = float64(b / 257)

			// Y (luma)
			yy = 0.299*r8 + 0.587*g8 + 0.114*b8
			y[j*width+i] = byte(yy)

			// Subsample U and V (4:2:0)
			if j%2 == 0 && i%2 == 0 {
				uIdx := (j/2)*(width/2) + i/2
				// SWAPPED: Assign Cr to u and Cb to v if that's what your decoder expects
				u[uIdx] = byte(0.5*r8 - 0.419*g8 - 0.081*b8 + 128)  // Cr
				v[uIdx] = byte(-0.169*r8 - 0.331*g8 + 0.5*b8 + 128) // Cb
			}
		}
	}

	data = nil

	return
}
