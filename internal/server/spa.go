package server

import (
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// spaHandler returns an http.Handler that serves static files from the given
// filesystem, falling back to index.html for client-side routing.
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(fsys))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cleanPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if cleanPath == "" {
			cleanPath = "index.html"
		}

		// Check if the requested file exists
		if f, err := fsys.Open(cleanPath); err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html directly for unmatched paths
		f, err := fsys.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		rs, ok := f.(io.ReadSeeker)
		if !ok {
			// fstest.MapFS files implement ReadSeeker
			content, err := io.ReadAll(f)
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(content)
			return
		}

		http.ServeContent(w, r, "index.html", stat.ModTime(), rs)
	})
}
