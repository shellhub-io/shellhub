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

// H264NVENCEncoder represents an H.264 encoder using NVIDIA NVENC hardware acceleration.
// It implements both the Encoder and HardwareAccelerator interfaces.
type H264NVENCEncoder struct {
	encoderID EncoderType
	codecCtx  *C.AVCodecContext
	frame     *C.AVFrame
	packet    *C.AVPacket
	frameNum  int64
	width     int
	height    int
}

var _ Encoder = (*H264NVENCEncoder)(nil)

// H264NVENCParams contains configuration parameters for NVIDIA NVENC encoder
type H264NVENCParams struct {
	Preset  string
	Tune    string
	Profile string
	// Bitrate in kbps.
	Bitrate uint
	// GOP (Group of Pictures) size - lower values improve interactivity.
	GOPSize int
	// B-frames - typically disabled for low-latency remote desktop.
	BFrames int
}

var DefaultH264NVENCParams = H264NVENCParams{
	Preset:  "ll",
	Tune:    "ull",
	Profile: "main",
	Bitrate: 500,
	GOPSize: 15,
	BFrames: 0,
}

// NewH264NVENCEncoder creates a new H.264 encoder using NVIDIA NVENC hardware acceleration.
func NewH264NVENCEncoder(width, height int, fps uint16, ops ...func(*H264NVENCParams)) (*H264NVENCEncoder, error) {
	params := DefaultH264NVENCParams
	for _, op := range ops {
		op(&params)
	}

	codec := C.avcodec_find_encoder_by_name(C.CString("h264_nvenc"))
	if codec == nil {
		return nil, errors.New("h264_nvenc encoder not found")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		return nil, errors.New("failed to allocate codec context")
	}

	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	ctx.framerate = C.AVRational{num: C.int(fps), den: 1}
	ctx.sample_aspect_ratio = C.AVRational{num: 1, den: 1}
	ctx.pix_fmt = C.AV_PIX_FMT_BGRA

	ctx.bit_rate = C.int64_t(params.Bitrate * 1000)
	ctx.rc_max_rate = C.int64_t((params.Bitrate * 1000) * 2)
	ctx.rc_min_rate = C.int64_t((params.Bitrate * 1000) / 2)
	ctx.gop_size = C.int(params.GOPSize)
	ctx.max_b_frames = C.int(params.BFrames)

	C.av_opt_set(unsafe.Pointer(ctx), C.CString("preset"), C.CString(params.Preset), 0)
	C.av_opt_set(unsafe.Pointer(ctx), C.CString("tune"), C.CString(params.Tune), 0)

	if C.avcodec_open2(ctx, codec, nil) < 0 {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to open H.264 NVENC codec")
	}

	frame := C.av_frame_alloc()
	if frame == nil {
		C.avcodec_free_context(&ctx)

		return nil, errors.New("failed to allocate software frame")
	}

	frame.width = C.int(width)
	frame.height = C.int(height)
	frame.format = C.AV_PIX_FMT_BGRA

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

	return &H264NVENCEncoder{
		encoderID: EncoderTypeH264,
		codecCtx:  ctx,
		frame:     frame,
		packet:    packet,
		width:     width,
		height:    height,
	}, nil
}

// Code returns the encoder type code.
func (e *H264NVENCEncoder) Code() EncoderType {
	return e.encoderID
}

// Encode encodes a frame using NVIDIA NVENC hardware acceleration.
func (e *H264NVENCEncoder) Encode(width, height uint16, data []byte) ([]byte, error) {
	if C.av_frame_make_writable(e.frame) < 0 {
		return nil, errors.New("frame not writable")
	}

	C.memcpy(unsafe.Pointer(e.frame.data[0]), unsafe.Pointer(&data[0]), C.size_t(len(data)))

	e.frame.pts = C.int64_t(e.frameNum)
	e.frameNum++

	if C.avcodec_send_frame(e.codecCtx, e.frame) < 0 {
		return nil, errors.New("failed to send frame to NVENC encoder")
	}

	ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
	if ret == C.AVERR_EAGAIN() || ret == C.AVERR_EOF() {
		return nil, nil
	}

	if ret < 0 {
		return nil, errors.New("failed to receive packet from NVENC")
	}

	out := C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size)
	C.av_packet_unref(e.packet)

	return out, nil
}

// Flush flushes any remaining encoded frames from the encoder buffer.
func (e *H264NVENCEncoder) Flush() ([][]byte, error) {
	var frames [][]byte
	C.avcodec_send_frame(e.codecCtx, nil)
	for {
		ret := C.avcodec_receive_packet(e.codecCtx, e.packet)
		if ret == C.AVERR_EOF() || ret == C.AVERR_EAGAIN() {
			break
		}

		if ret < 0 {
			return nil, errors.New("NVENC flush failed")
		}

		frames = append(frames, C.GoBytes(unsafe.Pointer(e.packet.data), e.packet.size))
		C.av_packet_unref(e.packet)
	}

	return frames, nil
}

// SupportsFormat checks if NVENC supports the given pixel format.
func (e *H264NVENCEncoder) SupportsFormat(format string) bool {
	return format == "NV12" || format == "BGRA"
}

// Close releases all resources associated with the encoder.
func (e *H264NVENCEncoder) Close() {
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
