package util

import (
	"net/http"
	"net/http/httptest"
)

type fakeBroker struct {
	srv *httptest.Server
	ch  chan *http.Request
}

func newFakeBroker() *fakeBroker {
	ch := make(chan *http.Request)
	hdl := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go func() {
			ch <- r
		}()
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(hdl)
	go srv.Start()
	return &fakeBroker{srv: srv, ch: ch}
}

func (f *fakeBroker) Close() {
	f.srv.Close()

}

func (f *fakeBroker) URLStr() string {
	return f.srv.URL
}

func (f *fakeBroker) recv(n int) []*http.Request {
	ret := make([]*http.Request, n)
	i := 0
	for req := range f.ch {
		ret[i] = req
		i++
	}
	return ret
}
