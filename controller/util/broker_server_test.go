package util

import (
	"net/http"
	"net/http/httptest"
)

type fakeBroker struct {
	srv *httptest.Server
}

func newFakeBroker() *fakeBroker {
	hdl := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(hdl)
	go srv.Start()
	return &fakeBroker{srv: srv}
}

func (f *fakeBroker) Close() {
	f.srv.Close()

}

func (f *fakeBroker) URLStr() string {
	return f.srv.URL
}
