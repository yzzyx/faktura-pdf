package main

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
// URL: * (matches any path)
// Used by: fallback for non-matching routes
// Required permissions: NONE
func FileServer(r chi.Router, path string, root http.FileSystem) error {
	if strings.ContainsAny(path, "{}*") {
		return errors.New("invalid fileserver path")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the client accepts compressed responses
		enc := r.Header.Get("Accept-Encoding")
		if strings.Contains(enc, "gzip") {
			// Cleanup requested path
			filename := r.URL.Path
			if !strings.HasPrefix(filename, "/") {
				filename = "/" + filename
			}
			filename = strings.TrimPrefix(filename, path[0:len(path)-1])
			filename = pathpkg.Clean(filename)

			// Check if a compressed version of the file exists
			f, err := root.Open(filename + ".gz")
			if err == nil {
				defer func() { _ = f.Close() }() // We don't care about the returnvalue from Close() here

				// We'll need to keep the original content-type still
				contentType := mime.TypeByExtension(filepath.Ext(filename))

				if contentType != "" {
					stat, err := f.Stat()
					if err != nil {
						fmt.Fprintf(os.Stderr, "cannot stat: %v\n", err)
						w.WriteHeader(http.StatusInternalServerError)
						return
					}

					// Set contenttype and encoding
					w.Header().Set("Content-Type", contentType)
					w.Header().Set("Content-Encoding", "gzip")

					// We don't know the content-length, so if it's set, remove it
					w.Header().Del("Content-Length")
					http.ServeContent(w, r, filename, stat.ModTime(), f)
					return
				}
			}
		}
		fs.ServeHTTP(w, r)
	}))
	return nil
}
