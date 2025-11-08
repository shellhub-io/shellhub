package encoders

/*
#cgo pkg-config: libavcodec libavutil libavdevice
#include <libavcodec/avcodec.h>
#include <libavutil/pixdesc.h>
#include <libavutil/opt.h>
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

// H264LibX264Encoder represents an H.264 encoder using libx264 software encoder.
// It implements the Encoder interface.
type H264LibX264Encoder struct {
	encoderID EncoderType
	codecCtx  *C.AVCodecContext
	frame     *C.AVFrame
	packet    *C.AVPacket
	frameNum  int64
	width     int
	height    int
}

var _ Encoder = (*H264LibX264Encoder)(nil)

// H264LibX264Params contains configuration parameters for libx264 encoder
type H264LibX264Params struct {
	Preset  string
	Tune    string
	Profile string
	// Bitrate in kbps.
	Bitrate uint
	// GOP (Group of Pictures) size - lower values improve interactivity.
	GOPSize int
	// Reference frames - lower values reduce latency.
	RefFrames int
	// B-frames - typically disabled for low-latency remote desktop.
	BFrames int
	// AQ (Adaptive Quantization) mode for better quality distribution.
	AQMode int
	// Lookahead - higher values improve quality but increase latency.
	Lookahead int
}

var DefaultH264LibX264Params = H264LibX264Params{
	Preset:    "veryfast",
	Tune:      "zerolatency",
	Profile:   "baseline",
	Bitrate:   500,
	GOPSize:   15,
	RefFrames: 1,
	BFrames:   0,
	AQMode:    2,
	Lookahead: 0,
}

// NewH264LibX264Encoder creates a new H.264 encoder using libx264 software encoder.
func NewH264LibX264Encoder(width, height int, fps uint16, ops ...func(*H264LibX264Params)) (*H264LibX264Encoder, error) {
	params := DefaultH264LibX264Params
	for _, op := range ops {
		op(&params)
	}

	codec := C.avcodec_find_encoder_by_name(C.CString("libx264"))
	if codec == nil {
		return nil, errors.New("libx264 encoder not found")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, errors.New("failed to allocate codec context")
	}

	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	ctx.framerate = C.AVRational{num: C.int(fps), den: 1}
	ctx.gop_size = C.int(params.GOPSize)
	ctx.max_b_frames = C.int(params.BFrames)
	ctx.pix_fmt = C.AV_PIX_FMT_NV12

	// Remote desktop optimal settings
	ctx.bit_rate = C.int64_t(params.Bitrate * 1000)
	ctx.rc_max_rate = C.int64_t((params.Bitrate * 1000) * 2)
	ctx.rc_min_rate = C.int64_t((params.Bitrate * 1000) / 2)
	ctx.refs = C.int(params.RefFrames)

	C.av_opt_set(ctx.priv_data, C.CString("preset"), C.CString(params.Preset), 0)
	C.av_opt_set(ctx.priv_data, C.CString("tune"), C.CString(params.Tune), 0)
	C.av_opt_set(ctx.priv_data, C.CString("profile"), C.CString(params.Profile), 0)
	C.av_opt_set_int(ctx.priv_data, C.CString("aq-mode"), C.int64_t(params.AQMode), 0)
	C.av_opt_set_int(ctx.priv_data, C.CString("rc-lookahead"), C.int64_t(params.Lookahead), 0)

	if C.avcodec_open2(ctx, codec, nil) < 0 {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to open H.264 libx264 codec")
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate software frame")
	}

	frame.width = C.int(width)
	frame.height = C.int(height)
	frame.format = C.AV_PIX_FMT_NV12

	if C.av_frame_get_buffer(frame, 32) < 0 {
		C.av_frame_free(&frame)
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate software frame buffer")
	}

	packet := C.av_packet_alloc()
	if packet == nil {
		C.av_frame_free(&frame)
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate packet")
	}

	return &H264LibX264Encoder{
		encoderID: EncoderTypeH264,
		codecCtx:  ctx,
		frame:     frame,
		packet:    packet,
		width:     width,
		height:    height,
	}, nil
}

// Code returns the encoder type code.
func (e *H264LibX264Encoder) Code() EncoderType {
	return e.encoderID
}

// Encode encodes a frame using libx264 software encoder.
func (e *H264LibX264Encoder) Encode(width, height uint16, data []byte) ([]byte, error) {
	y, uv := bgraToNV12(int(width), int(height), data)

	if C.av_frame_make_writable(e.frame) < 0 {
		return nil, errors.New("frame not writable")
	}

	C.memcpy(unsafe.Pointer(e.frame.data[0]), unsafe.Pointer(&y[0]), C.size_t(len(y)))
	C.memcpy(unsafe.Pointer(e.frame.data[1]), unsafe.Pointer(&uv[0]), C.size_t(len(uv)))

	e.frame.pts = C.int64_t(e.frameNum)
	e.frameNum++

	if C.avcodec_send_frame(e.codecCtx, e.frame) < 0 {
		return nil, errors.New("failed to send frame to libx264 encoder")
	}

	ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
	if ret == C.AVERR_EAGAIN() || ret == C.AVERR_EOF() {
		return nil, nil
	}

	if ret < 0 {
		return nil, errors.New("failed to receive packet from libx264")
	}

	out := C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size)
	C.av_packet_unref(e.packet)

	return out, nil
}

// Flush flushes any remaining encoded frames from the encoder buffer.
func (e *H264LibX264Encoder) Flush() ([][]byte, error) {
	var frames [][]byte
	C.avcodec_send_frame(e.codecCtx, nil)
	for {
		ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
		if ret == C.AVERR_EOF() || ret == C.AVERR_EAGAIN() {
			break
		}

		if ret < 0 {
			return nil, errors.New("libx264 flush failed")
		}

		frames = append(frames, C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size))
		C.av_packet_unref(e.packet)
	}

	return frames, nil
}

// Close releases all resources associated with the encoder.
func (e *H264LibX264Encoder) Close() {
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
