package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"html"
	"k8s.io/api/admission/v1beta1"
	"net/http"

	_ "k8s.io/apimachinery/pkg/runtime"
)

func main() {
	var tlsConfig TLSConfig
	tlsConfig.addFlags()
	flag.Parse()

	http.HandleFunc("/webhooks/admission/allow-all", allowAll)
	http.HandleFunc("/health/", health)

	if tlsConfig.KeyFile == "" || tlsConfig.CertFile == "" {
		serveHTTP(":8080")
	} else {
		serveHTTPS(":8080", tlsConfig)
	}
}

// TLSConfig is inspired by https://github.com/kubernetes/kubernetes/blob/release-1.9/test/images/webhook/main.go
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

func (c *TLSConfig) addFlags() {
	flag.StringVar(&c.CertFile, "tls-cert-file", c.CertFile, "")
	flag.StringVar(&c.KeyFile, "tls-private-key-file", c.KeyFile, "")
}


// allowAll allows any admission request
func allowAll(writer http.ResponseWriter, request *http.Request) {
	glog.Infof("got %q", html.EscapeString(request.URL.Path))

	admissionResponse := v1beta1.AdmissionResponse{}
	admissionResponse.Allowed = true

	admissionReview := v1beta1.AdmissionReview{}
	admissionReview.Response = &admissionResponse

	glog.Infof("sending response: %v", admissionReview)

	response, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Error(err)
	}
	if _, err := writer.Write(response); err != nil {
		glog.Error(err)
	}
}

// health replies Ok to every request
func health(writer http.ResponseWriter, request *http.Request) {
	glog.V(2).Infof("got %q", html.EscapeString(request.URL.Path))
	if _, err := fmt.Fprintf(writer, "Ok, %q", html.EscapeString(request.URL.Path)); err != nil {
		glog.Error(err)
	}
}

// serveHTTP starts an HTTP server listening on the specified address
func serveHTTP(addr string) {
	glog.Info("starting HTTP server...")

	server := &http.Server{
		Addr: addr,
	}
	err := server.ListenAndServe()
	if err != nil {
		glog.Fatalf("failed to start HTTP server: %s", err)
	}
}

// serveHTTP starts an HTTP server listening on the specified address with the given TLS config
func serveHTTPS(addr string, config TLSConfig) {
	glog.Info("starting HTTPS server...")

	serverCertificate, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		glog.Error(err)
	}

	server := &http.Server{
		Addr: addr,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{serverCertificate},
		},
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		glog.Fatalf("failed to start HTTPS server: %s", err)
	}
}
