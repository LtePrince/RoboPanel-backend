package camera

/*
#cgo LDFLAGS: -lrealsense2
#cgo CFLAGS: -I/usr/include/librealsense2

#include "rs.h"
#include "h/rs_pipeline.h"
#include "h/rs_context.h"
#include "h/rs_config.h"
#include "h/rs_frame.h"
#include "h/rs_sensor.h"
#include <stdlib.h>
#include <string.h>

// Helper: return error message or empty string.
static const char* rs2_err_str(rs2_error* e) {
    return e ? rs2_get_error_message(e) : "";
}
*/
import "C"

import (
	"fmt"
	"image"
	"image/jpeg"
	"bytes"
	"unsafe"
)

// Config mirrors Python CameraStreamConfig.
type Config struct {
	Serial      string
	Width       int
	Height      int
	FPS         int
	JpegQuality int
	TimeoutMs   int
}

func DefaultConfig() Config {
	return Config{
		Serial:      "919122072662",
		Width:       640,
		Height:      480,
		FPS:         30,
		JpegQuality: 70,
		TimeoutMs:   1000,
	}
}

// Reader wraps the librealsense2 pipeline.
type Reader struct {
	cfg      Config
	ctx      *C.rs2_context
	pipe     *C.rs2_pipeline
	running  bool
}

func NewReader(cfg Config) *Reader {
	return &Reader{cfg: cfg}
}

func (r *Reader) Start() error {
	var e *C.rs2_error

	r.ctx = C.rs2_create_context(C.RS2_API_VERSION, &e)
	if err := cgoErr("create_context", e); err != nil {
		return err
	}

	r.pipe = C.rs2_create_pipeline(r.ctx, &e)
	if err := cgoErr("create_pipeline", e); err != nil {
		C.rs2_delete_context(r.ctx)
		return err
	}

	cfg := C.rs2_create_config(&e)
	if err := cgoErr("create_config", e); err != nil {
		C.rs2_delete_pipeline(r.pipe)
		C.rs2_delete_context(r.ctx)
		return err
	}
	defer C.rs2_delete_config(cfg)

	if r.cfg.Serial != "" {
		serial := C.CString(r.cfg.Serial)
		defer C.free(unsafe.Pointer(serial))
		C.rs2_config_enable_device(cfg, serial, &e)
		if err := cgoErr("enable_device", e); err != nil {
			C.rs2_delete_pipeline(r.pipe)
			C.rs2_delete_context(r.ctx)
			return err
		}
	}

	C.rs2_config_enable_stream(
		cfg,
		C.RS2_STREAM_COLOR,
		-1,
		C.int(r.cfg.Width),
		C.int(r.cfg.Height),
		C.RS2_FORMAT_BGR8,
		C.int(r.cfg.FPS),
		&e,
	)
	if err := cgoErr("enable_stream", e); err != nil {
		C.rs2_delete_pipeline(r.pipe)
		C.rs2_delete_context(r.ctx)
		return err
	}

	profile := C.rs2_pipeline_start_with_config(r.pipe, cfg, &e)
	if err := cgoErr("pipeline_start", e); err != nil {
		C.rs2_delete_pipeline(r.pipe)
		C.rs2_delete_context(r.ctx)
		return err
	}
	C.rs2_delete_pipeline_profile(profile)

	r.running = true
	return nil
}

// ReadJPEG captures one color frame and returns JPEG bytes.
func (r *Reader) ReadJPEG() ([]byte, error) {
	if !r.running {
		return nil, fmt.Errorf("camera not started")
	}

	var e *C.rs2_error

	frameset := C.rs2_pipeline_wait_for_frames(r.pipe, C.uint(r.cfg.TimeoutMs), &e)
	if err := cgoErr("wait_for_frames", e); err != nil {
		return nil, err
	}
	defer C.rs2_release_frame(frameset)

	colorFrame := C.rs2_extract_frame(frameset, 0, &e)
	if err := cgoErr("extract_frame", e); err != nil {
		return nil, err
	}
	defer C.rs2_release_frame(colorFrame)

	dataPtr := C.rs2_get_frame_data(colorFrame, &e)
	if err := cgoErr("get_frame_data", e); err != nil {
		return nil, err
	}

	stride := r.cfg.Width * 3 // BGR8: 3 bytes per pixel
	size := stride * r.cfg.Height
	raw := C.GoBytes(unsafe.Pointer(dataPtr), C.int(size))

	return encodeJPEG(raw, r.cfg.Width, r.cfg.Height, r.cfg.JpegQuality)
}

func (r *Reader) Stop() {
	if !r.running {
		return
	}
	var e *C.rs2_error
	C.rs2_pipeline_stop(r.pipe, &e)
	C.rs2_delete_pipeline(r.pipe)
	C.rs2_delete_context(r.ctx)
	r.running = false
}

// encodeJPEG converts a BGR8 byte slice to JPEG.
// librealsense2 delivers BGR (blue first); JPEG encoder expects RGB.
func encodeJPEG(bgr []byte, width, height, quality int) ([]byte, error) {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			i := (y*width + x) * 3
			off := (y*width+x)*4
			img.Pix[off+0] = bgr[i+2] // R ← B slot
			img.Pix[off+1] = bgr[i+1] // G
			img.Pix[off+2] = bgr[i+0] // B ← R slot
			img.Pix[off+3] = 0xff
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("jpeg encode: %w", err)
	}
	return buf.Bytes(), nil
}

func cgoErr(op string, e *C.rs2_error) error {
	if e == nil {
		return nil
	}
	msg := C.GoString(C.rs2_err_str(e))
	C.rs2_free_error(e)
	return fmt.Errorf("librealsense2 %s: %s", op, msg)
}
