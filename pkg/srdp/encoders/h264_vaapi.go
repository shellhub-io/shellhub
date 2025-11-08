package encoders

/*
#cgo pkg-config: libavcodec libavutil libavdevice
#include <libavcodec/avcodec.h>
#include <libavutil/pixdesc.h>
#include <libavutil/hwcontext.h>
#include <libavutil/opt.h>
#include <errno.h>

// expose AVERROR_* macros for Go
static inline int AVERR_EOF() { return AVERROR_EOF; }
static inline int AVERR_EAGAIN() { return AVERROR(EAGAIN); }

static int set_hwframe_ctx(AVCodecContext *ctx, AVBufferRef *hw_device_ctx, int width, int height)
{
    AVBufferRef *hw_frames_ref;
    AVHWFramesContext *frames_ctx = NULL;
    int err = 0;

    if (!(hw_frames_ref = av_hwframe_ctx_alloc(hw_device_ctx))) {
        return -1;
    }

    frames_ctx = (AVHWFramesContext *)(hw_frames_ref->data);
    frames_ctx->format    = AV_PIX_FMT_VAAPI;
    frames_ctx->sw_format = AV_PIX_FMT_NV12;
    frames_ctx->width     = width;
    frames_ctx->height    = height;
    frames_ctx->initial_pool_size = 20;

    if ((err = av_hwframe_ctx_init(hw_frames_ref)) < 0) {
        av_buffer_unref(&hw_frames_ref);
        return err;
    }

    ctx->hw_frames_ctx = av_buffer_ref(hw_frames_ref);
    if (!ctx->hw_frames_ctx)
        err = AVERROR(ENOMEM);

    av_buffer_unref(&hw_frames_ref);
    return err;
}
*/
import "C"

import (
	"errors"
	"unsafe"
)

// H264VAAPIEncoder represents an H.264 encoder using VAAPI hardware acceleration.
// It implements both the Encoder and HardwareAccelerator interfaces.
type H264VAAPIEncoder struct {
	encoderID   EncoderType
	codecCtx    *C.AVCodecContext
	hwDeviceCtx *C.AVBufferRef
	swFrame     *C.AVFrame
	hwFrame     *C.AVFrame
	packet      *C.AVPacket
	frameNum    int64
	width       int
	height      int
}

var _ Encoder = (*H264VAAPIEncoder)(nil)

type H264VAAPIParams struct {
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
}

var DefaultH264VAAPIParams = H264VAAPIParams{
	Preset:    "veryfast",
	Tune:      "zerolatency",
	Profile:   "baseline",
	Bitrate:   500,
	GOPSize:   15,
	RefFrames: 1,
	BFrames:   0,
}

// NewH264VAAPIEncoder creates a new H.264 encoder using VAAPI hardware acceleration.
func NewH264VAAPIEncoder(width, height int, fps uint16, ops ...func(*H264VAAPIParams)) (*H264VAAPIEncoder, error) {
	params := DefaultH264VAAPIParams
	for _, op := range ops {
		op(&params)
	}

	var hwDeviceCtx *C.AVBufferRef
	if C.av_hwdevice_ctx_create(&hwDeviceCtx, C.AV_HWDEVICE_TYPE_VAAPI, nil, nil, 0) < 0 {
		return nil, errors.New("failed to create VAAPI device")
	}

	codec := C.avcodec_find_encoder_by_name(C.CString("h264_vaapi"))
	if codec == nil {
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("h264_vaapi encoder not found")
	}

	ctx := C.avcodec_alloc_context3(codec)
	if ctx == nil {
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate codec context")
	}

	ctx.width = C.int(width)
	ctx.height = C.int(height)
	ctx.time_base = C.AVRational{num: 1, den: C.int(fps)}
	ctx.framerate = C.AVRational{num: C.int(fps), den: 1}
	ctx.sample_aspect_ratio = C.AVRational{num: 1, den: 1}
	ctx.pix_fmt = C.AV_PIX_FMT_VAAPI

	ctx.bit_rate = C.int64_t(params.Bitrate * 1000)
	ctx.rc_max_rate = C.int64_t((params.Bitrate * 1000) * 2)
	ctx.rc_min_rate = C.int64_t((params.Bitrate * 1000) / 2)
	ctx.gop_size = C.int(params.GOPSize)
	ctx.max_b_frames = C.int(params.BFrames)
	ctx.refs = C.int(params.RefFrames)

	if C.set_hwframe_ctx(ctx, hwDeviceCtx, C.int(width), C.int(height)) < 0 {
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to set hwframe context")
	}

	if C.avcodec_open2(ctx, codec, nil) < 0 {
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to open H.264 codec")
	}

	swFrame := C.av_frame_alloc()
	if swFrame == nil {
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate software frame")
	}

	swFrame.width = C.int(width)
	swFrame.height = C.int(height)
	swFrame.format = C.AV_PIX_FMT_NV12

	if C.av_frame_get_buffer(swFrame, 32) < 0 {
		C.av_frame_free(&swFrame)
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate software frame buffer")
	}

	hwFrame := C.av_frame_alloc()
	if hwFrame == nil {
		C.av_frame_free(&swFrame)
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate hardware frame")
	}

	if C.av_hwframe_get_buffer(ctx.hw_frames_ctx, hwFrame, 0) < 0 {
		C.av_frame_free(&hwFrame)
		C.av_frame_free(&swFrame)
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate hardware frame buffer")
	}

	if hwFrame.hw_frames_ctx == nil {
		C.av_frame_free(&hwFrame)
		C.av_frame_free(&swFrame)
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("hw_frames_ctx is NULL")
	}

	packet := C.av_packet_alloc()
	if packet == nil {
		C.av_frame_free(&hwFrame)
		C.av_frame_free(&swFrame)
		C.avcodec_free_context(&ctx)
		C.av_buffer_unref(&hwDeviceCtx)

		return nil, errors.New("failed to allocate packet")
	}

	return &H264VAAPIEncoder{
		encoderID:   EncoderTypeH264,
		codecCtx:    ctx,
		hwDeviceCtx: hwDeviceCtx,
		swFrame:     swFrame,
		hwFrame:     hwFrame,
		packet:      packet,
		width:       width,
		height:      height,
	}, nil
}

// Code returns the encoder type code.
func (e *H264VAAPIEncoder) Code() EncoderType {
	return e.encoderID
}

// Encode encodes a frame using VAAPI hardware acceleration.
func (e *H264VAAPIEncoder) Encode(width, height uint16, data []byte) ([]byte, error) {
	y, uv := bgraToNV12(int(width), int(height), data)

	if C.av_frame_make_writable(e.swFrame) < 0 {
		return nil, errors.New("frame not writable")
	}

	C.memcpy(unsafe.Pointer(e.swFrame.data[0]), unsafe.Pointer(&y[0]), C.size_t(len(y)))
	C.memcpy(unsafe.Pointer(e.swFrame.data[1]), unsafe.Pointer(&uv[0]), C.size_t(len(uv)))

	if C.av_hwframe_transfer_data(e.hwFrame, e.swFrame, 0) < 0 {
		return nil, errors.New("failed to transfer data to hardware frame")
	}

	e.hwFrame.pts = C.int64_t(e.frameNum)
	e.frameNum++

	if C.avcodec_send_frame(e.codecCtx, e.hwFrame) < 0 {
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

// Flush flushes any remaining encoded frames from the encoder buffer.
func (e *H264VAAPIEncoder) Flush() ([][]byte, error) {
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

// SupportsFormat checks if VAAPI supports the given pixel format.
func (e *H264VAAPIEncoder) SupportsFormat(format string) bool {
	return format == "NV12"
}

// Close releases all resources associated with the encoder.
func (e *H264VAAPIEncoder) Close() {
	if e.packet != nil {
		C.av_packet_free(&e.packet)
	}

	if e.hwFrame != nil {
		C.av_frame_free(&e.hwFrame)
	}

	if e.swFrame != nil {
		C.av_frame_free(&e.swFrame)
	}

	if e.codecCtx != nil {
		C.avcodec_free_context(&e.codecCtx)
	}

	if e.hwDeviceCtx != nil {
		C.av_buffer_unref(&e.hwDeviceCtx)
	}
}
