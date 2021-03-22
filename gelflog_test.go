package traefik_gelf_plugin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpHandlerMock struct{}

func (h *httpHandlerMock) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestGoodConfig(t *testing.T) {
	config := CreateConfig()
	config.GelfEndpoint = "192.168.2.4"
	config.GelfPort = 12203
	g, err := New(nil, &httpHandlerMock{}, config, "GelfLogger")

	if err != nil {
		t.Fatal(err)
	}


	req := httptest.NewRequest(http.MethodGet, "http://localhost/some/path", nil)
	req.RemoteAddr = "4.0.0.0:34000"
	rw := httptest.NewRecorder()

	g.ServeHTTP(rw, req)

	if config.EmitTraceId {
		traceHeader := req.Header.Get(config.TraceIdHeader)
		if traceHeader == "" {
			t.Fatal("trace id empty")
		}
	}

	if config.EmitRequestStart {
		reqStartHeader := req.Header.Get(config.RequestStartTimeHeader)
		if reqStartHeader == "" {
			t.Fatal("request start empty")
		}
	}
}
func TestHostnameOverride(t *testing.T) {
	config := CreateConfig()
	config.GelfEndpoint = "192.168.2.4"
	config.GelfPort = 12203
	config.HostnameOverride = "avimac01.av.local"
	g, err := New(nil, &httpHandlerMock{}, config, "GelfLogger")

	if err != nil {
		t.Fatal(err)
	}


	req := httptest.NewRequest(http.MethodGet, "http://localhost/some/path", nil)
	req.RemoteAddr = "4.0.0.0:34000"
	req.Header.Add("X-Test-Redundant", "abc")
	req.Header.Add("X-Test-Redundant", "123")
	rw := httptest.NewRecorder()

	g.ServeHTTP(rw, req)

	if config.EmitTraceId {
		traceHeader := req.Header.Get(config.TraceIdHeader)
		if traceHeader == "" {
			t.Fatal("trace id empty")
		}
	}

	if config.EmitRequestStart {
		reqStartHeader := req.Header.Get(config.RequestStartTimeHeader)
		if reqStartHeader == "" {
			t.Fatal("request start empty")
		}
	}
}
func TestBadConfig(t *testing.T) {
	config := CreateConfig()

	g, err := New(nil, &httpHandlerMock{}, config, "GelfLogger")

	if g != nil {
		t.Fatal(err)
	}

}
func TestMissingPort(t *testing.T) {
	config := CreateConfig()
	config.GelfEndpoint = "192.168.2.4"
	g, err := New(nil, &httpHandlerMock{}, config, "GelfLogger")

	if g != nil {
		t.Fatal(err)
	}
}
func TestMissingConfig(t *testing.T) {
	g, err := New(nil, &httpHandlerMock{}, nil, "GelfLogger")
	if g != nil {
		t.Fatal(err)
	}
}