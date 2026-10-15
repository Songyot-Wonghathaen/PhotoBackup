//go:build windows

package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"syscall"
	"unsafe"
)

var (
	shell32                         = syscall.NewLazyDLL("shell32.dll")
	ole32                           = syscall.NewLazyDLL("ole32.dll")
	gdi32                           = syscall.NewLazyDLL("gdi32.dll")
	user32                          = syscall.NewLazyDLL("user32.dll")
	procSHCreateItemFromParsingName = shell32.NewProc("SHCreateItemFromParsingName")
	procCoInitializeEx              = ole32.NewProc("CoInitializeEx")
	procCoUninitialize              = ole32.NewProc("CoUninitialize")
	procDeleteObject                = gdi32.NewProc("DeleteObject")
	procGetDC                       = user32.NewProc("GetDC")
	procReleaseDC                   = user32.NewProc("ReleaseDC")
	procGetDIBits                   = gdi32.NewProc("GetDIBits")
	procGetObject                   = gdi32.NewProc("GetObjectW")
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var iidIShellItemImageFactory = guid{
	Data1: 0xbcc18b79,
	Data2: 0xba16,
	Data3: 0x442f,
	Data4: [8]byte{0x80, 0xc4, 0x8a, 0x59, 0xc3, 0x0c, 0x46, 0x3b},
}

type winSize struct {
	CX int32
	CY int32
}

type winBitmap struct {
	Type       int32
	Width      int32
	Height     int32
	WidthBytes int32
	Planes     uint16
	BitsPixel  uint16
	Bits       uintptr
}

type winBitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

// GetThumbnailBytes extracts a high-quality JPEG thumbnail of any image (JPG, PNG, HEIC, RAW) using Windows Shell
func GetThumbnailBytes(filePath string, maxDim int) ([]byte, error) {
	procCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED
	defer procCoUninitialize.Call()

	pathPtr, err := syscall.UTF16PtrFromString(filePath)
	if err != nil {
		return nil, err
	}

	var pFactory uintptr
	hr, _, _ := procSHCreateItemFromParsingName.Call(
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		uintptr(unsafe.Pointer(&iidIShellItemImageFactory)),
		uintptr(unsafe.Pointer(&pFactory)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("SHCreateItem failed: 0x%x", hr)
	}
	defer func() {
		vtable := *(**[3]uintptr)(unsafe.Pointer(pFactory))
		syscall.SyscallN(vtable[2], pFactory) // Release
	}()

	vtable := *(**[4]uintptr)(unsafe.Pointer(pFactory))
	getImageProc := vtable[3]

	var hBitmap uintptr
	size := winSize{CX: int32(maxDim), CY: int32(maxDim)}
	flags := uint32(0x0) // SIIGBF_RESIZETOFIT

	hr, _, _ = syscall.SyscallN(
		getImageProc,
		pFactory,
		*(*uintptr)(unsafe.Pointer(&size)),
		uintptr(flags),
		uintptr(unsafe.Pointer(&hBitmap)),
	)
	if hr != 0 {
		return nil, fmt.Errorf("GetImage failed: 0x%x", hr)
	}
	defer procDeleteObject.Call(hBitmap)

	var bm winBitmap
	procGetObject.Call(hBitmap, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
	w, h := int(bm.Width), int(bm.Height)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid bitmap dimensions: %dx%d", w, h)
	}

	hdc, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, hdc)

	var bi winBitmapInfoHeader
	bi.Size = uint32(unsafe.Sizeof(bi))
	bi.Width = int32(w)
	bi.Height = -int32(h) // Top-down
	bi.Planes = 1
	bi.BitCount = 32
	bi.Compression = 0 // BI_RGB

	buf := make([]byte, w*h*4)
	ret, _, _ := procGetDIBits.Call(
		hdc,
		hBitmap,
		0,
		uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bi)),
		0, // DIB_RGB_COLORS
	)
	if ret == 0 {
		return nil, fmt.Errorf("GetDIBits failed")
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			offset := (y*w + x) * 4
			b := buf[offset]
			g := buf[offset+1]
			r := buf[offset+2]
			a := buf[offset+3]
			if a == 0 {
				a = 255
			}
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: a})
		}
	}

	var out bytes.Buffer
	err = jpeg.Encode(&out, img, &jpeg.Options{Quality: 80})
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
