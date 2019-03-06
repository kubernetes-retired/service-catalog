package brokerapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

type httpRouter struct {
	muxRouter *mux.Router
}

func newHttpRouter() httpRouter {
	return httpRouter{
		muxRouter: mux.NewRouter(),
	}
}

func (httpRouter httpRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	httpRouter.muxRouter.ServeHTTP(w, req)
}

func (httpRouter httpRouter) Get(url string, handler http.HandlerFunc) {
	httpRouter.muxRouter.HandleFunc(url, handler).Methods("GET")
}

func (httpRouter httpRouter) Put(url string, handler http.HandlerFunc) {
	httpRouter.muxRouter.HandleFunc(url, handler).Methods("PUT")
}

func (httpRouter httpRouter) Delete(url string, handler http.HandlerFunc) {
	httpRouter.muxRouter.HandleFunc(url, handler).Methods("DELETE")
}

func (httpRouter) Vars(req *http.Request) map[string]string {
	return mux.Vars(req)
}
