package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/quic-go/quic-go/http3"
)

type QUICRoundTripper struct {
	http3RoundTripper *http3.RoundTripper
	http3RoundTripOpt http3.RoundTripOpt
}

type RoundRobinReverseProxy struct {
	backends []*httputil.ReverseProxy
	current  uint64
	mutex    sync.Mutex
}

func (qrt QUICRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return qrt.http3RoundTripper.RoundTripOpt(req, qrt.http3RoundTripOpt)
}

func NewProxy(target *url.URL, useQUIC bool) *httputil.ReverseProxy {
	fmt.Printf("Proxying to %s, QUIC: %t\n", target.Host, useQUIC)

	rp := httputil.NewSingleHostReverseProxy(target)

	director := rp.Director
	rp.Director = func(req *http.Request) {
		director(req)
		req.Host = target.Host
	}

	if useQUIC {

		roundTripper := &QUICRoundTripper{
			http3RoundTripper: &http3.RoundTripper{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			http3RoundTripOpt: http3.RoundTripOpt{},
		}
		rp.Transport = roundTripper

	}
	return rp
}

func (p *RoundRobinReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	backend := p.backends[p.current%uint64(len(p.backends))]
	p.current++
	backend.ServeHTTP(w, r)
}

func NewRoundRobinReverseProxy(targets []string, useQuic bool) *RoundRobinReverseProxy {
	backends := make([]*httputil.ReverseProxy, len(targets))
	for i, target := range targets {
		url, err := url.Parse(target)
		if err != nil {
			log.Fatalf("Erro ao analisar o destino %s: %s", target, err)
		}
		backends[i] = NewProxy(url, useQuic)
	}
	return &RoundRobinReverseProxy{backends: backends}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	upstream := os.Getenv("SEND_UPSTREAM")
	listenAddr := ":" + os.Getenv("LISTEN_PORT")
	useQUIC := strings.ToUpper(os.Getenv("QUIC")) == "TRUE"

	proxy := NewRoundRobinReverseProxy(strings.Split(upstream, ","), useQUIC)

	http.HandleFunc("/", proxy.ServeHTTP)
	srv := &http.Server{
		Addr: listenAddr,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}
