package server

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"PhotoVault/repository"

	"gorm.io/gorm"

	// Register pure Go image decoders
	_ "image/gif"
	_ "image/png"
)

// thumbMaxLongSide is the target long side (in pixels) of thumbnails
const thumbMaxLongSide = 400

// ImageHandler serves original images (/image/{id}) and cached thumbnails
// (/thumb/{id}). Any other request is delegated to the legacy handler.
type ImageHandler struct {
	db        *gorm.DB
	photoRepo repository.GalleryRepository
	fallback  http.Handler
}

// NewHandler wraps the legacy thumbnail handler with id-based image routes.
// Requests outside /image/ and /thumb/ are delegated to fallback.
func NewHandler(db *gorm.DB, fallback http.Handler) http.Handler {
	return &ImageHandler{
		db:        db,
		photoRepo: repository.NewGalleryRepository(db),
		fallback:  fallback,
	}
}

func (h *ImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/image/"):
		h.serveByID(w, r, "/image/", false)
	case strings.HasPrefix(r.URL.Path, "/thumb/"):
		h.serveByID(w, r, "/thumb/", true)
	default:
		if h.fallback != nil {
			h.fallback.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

// serveByID resolves a photo by numeric id and serves its original file or thumbnail
func (h *ImageHandler) serveByID(w http.ResponseWriter, r *http.Request, prefix string, asThumb bool) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, prefix)
	if idText == "" || strings.ContainsAny(idText, `/\`) {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(idText, 10, 32)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}

	if h.db == nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	photo, err := h.photoRepo.FindActiveByID(r.Context(), int32(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "failed to look up photo", http.StatusInternalServerError)
		return
	}

	// Compose the file path and make sure it stays inside the destination folder
	fullPath := filepath.Join(photo.DestinationPath, photo.FileName)
	if !isUnderRoot(photo.DestinationPath, fullPath) {
		http.NotFound(w, r)
		return
	}

	if asThumb {
		serveThumb(w, r, photo.ID, fullPath)
		return
	}
	serveOriginal(w, r, fullPath)
}

// serveOriginal streams the original file with an explicit Content-Type
func serveOriginal(w http.ResponseWriter, r *http.Request, fullPath string) {
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", contentTypeFor(fullPath))
	http.ServeFile(w, r, fullPath)
}

// serveThumb serves the cached thumbnail when available, otherwise shrinks the
// image with pure Go decoders and caches the result. If the image cannot be
// shrunk (undecodable format), the original file is served as fallback.
func serveThumb(w http.ResponseWriter, r *http.Request, id int32, fullPath string) {
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}

	if cachePath, ok := thumbCachePath(id, info); ok {
		if cached, cerr := os.Stat(cachePath); cerr == nil && !cached.IsDir() {
			w.Header().Set("Content-Type", "image/jpeg")
			http.ServeFile(w, r, cachePath)
			return
		}
	}

	data, resized, err := shrinkJPEG(fullPath, thumbMaxLongSide)
	if err != nil || !resized {
		// Undecodable, or already within the target size: serve the original
		serveOriginal(w, r, fullPath)
		return
	}

	if cachePath, ok := thumbCachePath(id, info); ok {
		_ = writeCacheFile(cachePath, data)
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// shrinkJPEG decodes an image with the pure Go stdlib decoders and re-encodes it
// as JPEG with the long side scaled down to maxLongSide. The resized return is
// false when the image is already within maxLongSide or cannot be decoded.
func shrinkJPEG(srcPath string, maxLongSide int) (data []byte, resized bool, err error) {
	f, err := os.Open(srcPath)
	if err != nil {
		return nil, false, err
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		return nil, false, err
	}

	bounds := src.Bounds()
	sw, sh := bounds.Dx(), bounds.Dy()
	if sw <= 0 || sh <= 0 {
		return nil, false, fmt.Errorf("invalid image dimensions: %dx%d", sw, sh)
	}

	longSide := sw
	if sh > longSide {
		longSide = sh
	}
	if longSide <= maxLongSide {
		// Already small enough: no resize needed
		return nil, false, nil
	}

	dw := sw * maxLongSide / longSide
	dh := sh * maxLongSide / longSide
	if dw < 1 {
		dw = 1
	}
	if dh < 1 {
		dh = 1
	}

	dst := resizeAreaAverage(src, dw, dh)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, false, err
	}
	return buf.Bytes(), true, nil
}

// resizeAreaAverage downscales src to dstW x dstH by averaging source pixels
// of each destination box. Transparent pixels are composited over white.
func resizeAreaAverage(src image.Image, dstW, dstH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	sb := src.Bounds()

	for y := 0; y < dstH; y++ {
		sy0 := sb.Min.Y + y*sb.Dy()/dstH
		sy1 := sb.Min.Y + (y+1)*sb.Dy()/dstH
		if sy1 <= sy0 {
			sy1 = sy0 + 1
		}
		for x := 0; x < dstW; x++ {
			sx0 := sb.Min.X + x*sb.Dx()/dstW
			sx1 := sb.Min.X + (x+1)*sb.Dx()/dstW
			if sx1 <= sx0 {
				sx1 = sx0 + 1
			}

			var r, g, b, a float64
			n := 0.0
			for yy := sy0; yy < sy1; yy++ {
				for xx := sx0; xx < sx1; xx++ {
					cr, cg, cb, ca := src.At(xx, yy).RGBA()
					r += float64(cr)
					g += float64(cg)
					b += float64(cb)
					a += float64(ca)
					n++
				}
			}
			if n == 0 {
				n = 1
			}

			am := a / n / 65535.0
			var rf, gf, bf float64
			if am > 0 {
				rf = (r / n / 65535.0) / am
				gf = (g / n / 65535.0) / am
				bf = (b / n / 65535.0) / am
			}
			// Composite over a white background (JPEG has no alpha channel)
			rf = rf*am + (1 - am)
			gf = gf*am + (1 - am)
			bf = bf*am + (1 - am)

			dst.SetRGBA(x, y, color.RGBA{
				R: clampByte(rf),
				G: clampByte(gf),
				B: clampByte(bf),
				A: 255,
			})
		}
	}
	return dst
}

func clampByte(v float64) uint8 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 255
	}
	return uint8(v * 255.0)
}

// thumbCacheDir returns the on-disk cache directory inside os.UserCacheDir()
func thumbCacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "PhotoVault", "thumbs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// thumbCachePath builds a cache file name that invalidates when the source changes
func thumbCachePath(id int32, info os.FileInfo) (string, bool) {
	dir, err := thumbCacheDir()
	if err != nil {
		return "", false
	}
	name := fmt.Sprintf("thumb_%d_%d_%d_%d.jpg", id, thumbMaxLongSide, info.Size(), info.ModTime().UnixNano())
	return filepath.Join(dir, name), true
}

// writeCacheFile stores data at cachePath atomically (temp file + rename)
func writeCacheFile(cachePath string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(cachePath), ".thumb-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, cachePath); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return nil
}

// isUnderRoot verifies that target is located inside root (path traversal guard)
func isUnderRoot(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	// Windows and macOS filesystems are case-insensitive
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		rootAbs = strings.ToLower(rootAbs)
		targetAbs = strings.ToLower(targetAbs)
	}
	return strings.HasPrefix(targetAbs, rootAbs)
}

// contentTypeFor maps a file extension to its HTTP content type
func contentTypeFor(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".bmp":
		return "image/bmp"
	case ".heic", ".heif":
		return "image/heic"
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
		return "application/octet-stream"
	}
}
