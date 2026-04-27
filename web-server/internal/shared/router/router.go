package router

import "net/http"

// Registrar is implemented by feature handlers that own a set of routes
// and wire them into the given mux.
type Registrar interface {
	Register(mux *http.ServeMux)
}

func New(registrars ...Registrar) *http.ServeMux {
	mux := http.NewServeMux()
	for _, r := range registrars {
		r.Register(mux)
	}
	return mux
}
