package httpx

import (
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type SPAHandler struct {
	fs        fs.FS
	indexPath string
}

func NewSPAHandler(dir string) (http.Handler, error) {
	if dir == "" {
		return nil, errors.New("static dir is required")
	}

	return &SPAHandler{
		fs:        os.DirFS(dir),
		indexPath: "index.html",
	}, nil
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}

	reqPath := r.URL.Path
	if reqPath == "" {
		reqPath = "/"
	}

	clean := path.Clean(reqPath)
	if clean == "." {
		clean = "/"
	}
	rel := strings.TrimPrefix(clean, "/")

	if rel != "" {
		if ok, isDir := h.exists(rel); ok && !isDir {
			h.serveFile(w, r, rel)
			return
		}
		if looksLikeAsset(rel) {
			http.NotFound(w, r)
			return
		}
	}

	h.serveFile(w, r, h.indexPath)
}

func (h *SPAHandler) exists(name string) (bool, bool) {
	st, err := fs.Stat(h.fs, name)
	if err != nil {
		return false, false
	}
	return true, st.IsDir()
}

func (h *SPAHandler) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	f, err := h.fs.Open(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()

	st, err := fs.Stat(h.fs, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if st.IsDir() {
		http.NotFound(w, r)
		return
	}

	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	if ct := mime.TypeByExtension(path.Ext(name)); ct != "" {
		w.Header().Set("Content-Type", ct)
	}

	rs, ok := f.(io.ReadSeeker)
	if !ok {
		http.NotFound(w, r)
		return
	}

	http.ServeContent(w, r, name, modTimeOrNow(st.ModTime()), rs)
}

func modTimeOrNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

func looksLikeAsset(rel string) bool {
	base := path.Base(rel)
	return strings.Contains(base, ".")
}
