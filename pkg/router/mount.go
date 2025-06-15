package router

import (
	"net/http"
	"strings"
)

func Mount(mux *http.ServeMux, prefix string, handler http.Handler) {
	p := strings.TrimRight(prefix, "/")
	
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == p {
			originalPath := r.URL.Path
			r.URL.Path = p + "/"
			http.StripPrefix(p, handler).ServeHTTP(w, r)
			r.URL.Path = originalPath
			return
		}
		http.StripPrefix(p, handler).ServeHTTP(w, r)
	})
	
	mux.Handle(p, wrapper)
	mux.Handle(p+"/", wrapper)
}
