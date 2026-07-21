package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chashma/lms/internal/platform/web"
)

// maxUploadBytes caps a single upload (videos). Mirrors the Java 512MB limit.
const maxUploadBytes = 512 << 20

// allowedExts is the per-kind extension allow-list (no executables/HTML).
var allowedExts = map[string]map[string]bool{
	"video": {"mp4": true, "webm": true, "mov": true, "m4v": true},
	"image": {"jpg": true, "jpeg": true, "png": true, "webp": true, "gif": true},
}

// uploads is the local-disk media upload endpoint. Swapping to S3 later would
// only change this file; the returned URL contract is preserved.
type uploads struct {
	root      string
	publicURL string
}

func newUploads(dir, publicURL string) (*uploads, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, err
	}
	return &uploads{root: root, publicURL: strings.TrimRight(publicURL, "/")}, nil
}

func (u *uploads) upload(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind == "" {
		kind = "video"
	}
	exts, ok := allowedExts[kind]
	if !ok {
		web.ErrorResponse(w, http.StatusBadRequest, "kind must be one of video, image")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		web.ErrorResponse(w, http.StatusBadRequest, "file must be provided")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		web.ErrorResponse(w, http.StatusBadRequest, "file must be provided")
		return
	}
	defer file.Close()

	ext := extension(header.Filename)
	if !exts[ext] {
		web.ErrorResponse(w, http.StatusBadRequest,
			fmt.Sprintf("unsupported %s file type .%s (allowed: %s)", kind, ext, strings.Join(keys(exts), ", ")))
		return
	}

	now := time.Now()
	subdir := fmt.Sprintf("%d/%02d", now.Year(), now.Month())
	dir := filepath.Join(u.root, subdir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		web.ServerError(w, r, err)
		return
	}
	name := randomName() + "." + ext
	dst, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		web.ServerError(w, r, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		web.ServerError(w, r, err)
		return
	}

	url := fmt.Sprintf("%s/uploads/%s/%s", u.publicURL, subdir, name)
	web.WriteJSON(w, http.StatusCreated, web.Envelope{"url": url, "filename": header.Filename}, nil)
}

// serve returns a static file handler for previously uploaded files.
func (u *uploads) serve() http.Handler {
	return http.StripPrefix("/uploads/", http.FileServer(http.Dir(u.root)))
}

func extension(filename string) string {
	dot := strings.LastIndex(filename, ".")
	if dot < 0 {
		return ""
	}
	return strings.ToLower(filename[dot+1:])
}

func randomName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
