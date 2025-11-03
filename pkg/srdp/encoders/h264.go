package encoders

/*
#cgo pkg-config: libavcodec libavutil
#include <libavcodec/avcodec.h>
#include <libavutil/opt.h>
#include <libavutil/imgutils.h>
#include <errno.h>

// expose AVERROR_* macros for Go
static inline int AVERR_EOF() { return AVERROR_EOF; }
static inline int AVERR_EAGAIN() { return AVERROR(EAGAIN); }
*/
import "C"

import (
	"errors"
	"unsafe"
)

type H264 struct {
	encoderID EncoderType
	codecCtx  *C.AVCodecContext
	frame     *C.AVFrame
	packet    *C.AVPacket
	frameNum  int64
	width     int
	height    int
}

type H264Params struct {
	Preset  string
	Tune    string
	Profile string
	// Bitrate in kbps.
	Bitrate uint
}

var DefaultH264Params = H264Params{
	Preset:  "veryfast",
	Tune:    "zerolatency",
	Profile: "baseline",
	Bitrate: 500,
}

func NewH264(width, height int, fps uint16, ops ...func(*H264Params)) (*H264, error) {
	params := DefaultH264Params
	for _, op := range ops {
		op(&params)
	}

	codec := C.avcodec_find_encoder(C.AV_CODEC_ID_H264)
	if codec == nil {
		return nil, errors.New("VAAPI encoder not found (h264_vaapi missing in ffmpeg build)")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, errors.New("failed to allocate codec context")
	}

	ctx.bit_rate = C.int64_t(params.Bitrate * 1000) // in bits per second // 1000 * 1000 = 1 Mbps
	ctx.rc_max_rate = C.int64_t((params.Bitrate * 1000) * 2)
	ctx.rc_min_rate = C.int64_t((params.Bitrate * 1000) / 2)

	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	ctx.framerate = C.AVRational{num: C.int(fps), den: 1}
	ctx.gop_size = C.int(fps)
	ctx.max_b_frames = 0
	ctx.pix_fmt = C.AV_PIX_FMT_YUV420P

	// Preset, tune, profile
	C.av_opt_set(ctx.priv_data, C.CString("preset"), C.CString(params.Preset), 0)
	C.av_opt_set(ctx.priv_data, C.CString("tune"), C.CString(params.Tune), 0)
	C.av_opt_set(ctx.priv_data, C.CString("profile"), C.CString(params.Profile), 0)

	if C.avcodec_open2(ctx, codec, nil) < 0 {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to open H.264 codec")
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate frame")
	}

	frame.format = C.int(ctx.pix_fmt)
	frame.width = C.int(width)
	frame.height = C.int(height)

	if C.av_frame_get_buffer(frame, 32) < 0 {
		C.av_frame_free(&frame)
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate frame buffer")
	}

	packet := C.av_packet_alloc()
	if packet == nil {
		C.av_frame_free(&frame)
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate packet")
	}

	return &H264{
		encoderID: EncoderTypeH264,
		codecCtx:  ctx,
		frame:     frame,
		packet:    packet,
		width:     width,
		height:    height,
	}, nil
}

// Code returns the encoder type code.
func (e *H264) Code() EncoderType {
	return e.encoderID
}

func (e *H264) Encode(width, height uint16, data []byte) ([]byte, error) {
	y, u, v := rgbToYUV420(int(width), int(height), data)

	if C.av_frame_make_writable(e.frame) < 0 {
		return nil, errors.New("frame not writable")
	}

	C.memcpy(unsafe.Pointer(e.frame.data[0]), unsafe.Pointer(&y[0]), C.size_t(len(y)))
	C.memcpy(unsafe.Pointer(e.frame.data[1]), unsafe.Pointer(&u[0]), C.size_t(len(u)))
	C.memcpy(unsafe.Pointer(e.frame.data[2]), unsafe.Pointer(&v[0]), C.size_t(len(v)))

	e.frame.pts = C.int64_t(e.frameNum)
	e.frameNum++

	// NOTE: This is the main encoding process. The idea is send a frame to the encoder and receive a packet, data
	// encoded in H.264 format, to be sent to the client.
	if C.avcodec_send_frame(e.codecCtx, e.frame) < 0 {
		return nil, errors.New("failed to send frame to encoder")
	}

	ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
	if ret == C.AVERR_EAGAIN() || ret == C.AVERR_EOF() {
		return nil, nil
	}

	if ret < 0 {
		return nil, errors.New("failed to receive packet")
	}

	out := C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size)
	C.av_packet_unref(e.packet)

	return out, nil
}

func (e *H264) Flush() ([][]byte, error) {
	var frames [][]byte
	C.avcodec_send_frame(e.codecCtx, nil)
	for {
		ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
		if ret == C.AVERR_EOF() || ret == C.AVERR_EAGAIN() {
			break
		}

		if ret < 0 {
			return nil, errors.New("flush failed")
		}

		frames = append(frames, C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size))
		C.av_packet_unref(e.packet)
	}

	return frames, nil
}

func (e *H264) Close() {
	if e.packet != nil {
		C.av_packet_free(&e.packet)
	}

	if e.frame != nil {
		C.av_frame_free(&e.frame)
	}

	if e.codecCtx != nil {
		C.avcodec_free_context(&e.codecCtx)
	}
}

// rgbToYUV420 converts RGBA 16-bit data to YUV420 planar format.
// TODO: Optimize this function for performance.
// TODO: Verify color accuracy.
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
