package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	env "github.com/caarlos0/env/v11"
)

type Config struct {
	EncodedHeader string        `env:"ENCODED_HEADER" envDefault:""            envExpand:"true"`
	SSLCert       string        `env:"SSL_CERT"       envDefault:"example.crt" envExpand:"true"`
	SSLKey        string        `env:"SSL_KEY"        envDefault:"example.key" envExpand:"true"`
	ListenAddress string        `env:"LISTEN_ADDRESS" envDefault:":8080"       envExpand:"true"`
	Timeout       time.Duration `env:"TIMEOUT"        envDefault:"3s"          envExpand:"true"`
}

func copyCloseWriter(dst io.WriteCloser, src io.Reader) {
	defer dst.Close()

	//nolint:errcheck // eat err, cause cannot response on hijacked connection
	_, _ = io.Copy(dst, src)
}

func proxy(responseWriter http.ResponseWriter, request *http.Request) {
	dialer := net.Dialer{}

	outgoingConn, err := dialer.DialContext(request.Context(), "tcp", request.Host)
	if err != nil {
		http.Error(responseWriter, request.Host+" unavailable", http.StatusServiceUnavailable)

		return
	}

	responseWriter.WriteHeader(http.StatusOK)

	hijacker, ok := responseWriter.(http.Hijacker)
	if !ok {
		http.Error(responseWriter, "hijacking not supported", http.StatusInternalServerError)

		return
	}

	incomeConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(responseWriter, "failed to hijack", http.StatusInternalServerError)

		return
	}

	go copyCloseWriter(outgoingConn, incomeConn)
	go copyCloseWriter(incomeConn, outgoingConn)
}

func handler(expectedValue string) http.HandlerFunc {
	if expectedValue == "" {
		return proxy
	}

	return http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		value, found := strings.CutPrefix(request.Header.Get("Proxy-Authorization"), "Basic ")
		if !found || value != expectedValue {
			http.Error(responseWriter, "", http.StatusProxyAuthRequired)

			return
		}

		proxy(responseWriter, request)
	})
}

func main() {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err)
	}

	server := http.Server{
		Addr:         cfg.ListenAddress,
		Handler:      handler(cfg.EncodedHeader),
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		// http1 only, because of the lack of support hijack in http2
		TLSNextProto: map[string]func(*http.Server, *tls.Conn, http.Handler){},
	}

	log.Fatal(server.ListenAndServeTLS(cfg.SSLCert, cfg.SSLKey))
}
