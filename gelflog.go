package traefik_gelf_plugin

import (
	"context"
	"fmt"
	"github.com/kjk/betterguid"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"net/http"
	"os"
	"time"
)

// Config holds the plugin configuration.
type Config struct {
	GelfEndpoint string `json:"gelfEndpoint,omitempty"`
	GelfPort int `json:"gelfPort,omitempty"`
	HostnameOverride string `json:"hostnameOverride"`
	EmitTraceId bool `json:"emitTraceId"`
	TraceIdHeader string `json:"traceIdHeader"`
	EmitRequestStart bool `json:"emitRequestStart"`
	RequestStartTimeHeader string `json:"requestStartTimeHeader"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{
		EmitTraceId: true,
		TraceIdHeader: "X-TraceId-AV",
		EmitRequestStart: true,
		RequestStartTimeHeader: "X-Request-Start",
	}
}

type GelfLog struct {
	Name       string
	Next       http.Handler
	Config     *Config
	GelfHostname	string
	GelfWriter *gelf.UDPWriter
}

// New creates and returns a plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	tLog := &GelfLog{
		Name: name,
		Next: next,
		Config: config,
	}
	if config == nil {
		return nil, fmt.Errorf("config can not be empty")
	}
	if config.GelfEndpoint == "" {
		return nil, fmt.Errorf("you must specify a GELF compatibile endpoint")
	}

	if config.GelfPort == 0 || config.GelfPort > 65353 {
		return nil, fmt.Errorf("you must specify a valid port")
	}

	if config.HostnameOverride == "" {
		tLog.GelfHostname, _ = os.Hostname()
	} else {
		tLog.GelfHostname = config.HostnameOverride
	}
	tLog.GelfWriter, _ = gelf.NewUDPWriter(fmt.Sprintf("%s:%d", config.GelfEndpoint, config.GelfPort))
	return tLog, nil
}


func (h *GelfLog) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := betterguid.New()
	if h.Config.EmitRequestStart {
		req.Header.Set(h.Config.RequestStartTimeHeader, fmt.Sprint(makeTimestampMilli()))
	}
	if h.Config.EmitTraceId {
		req.Header.Set(h.Config.TraceIdHeader, fmt.Sprint(id))
	}
	if h.GelfWriter != nil {
		var headerMap = map[string]interface{}{}
		for str, val := range req.Header {
			for index, iVal := range val {
				headerName := str
				if index > 0 {
					headerName = fmt.Sprintf("%s_%d", str, index)
				}
				headerMap[headerName] = iVal
			}
		}
		headerMap["Host"] = req.Host
		message := wrapMessage(fmt.Sprintf("Request to %s", req.Host), fmt.Sprintf("Request to %s", req.Host), h.GelfHostname, headerMap)
		h.GelfWriter.WriteMessage(message)
	}
	h.Next.ServeHTTP(rw, req)
}

func wrapMessage(s string, f string, h string, ex map[string]interface{}) *gelf.Message {
	/*
		Level is a standard syslog level
		Facility is deprecated
		Line is deprecated
		File is deprecated
	*/

	m := &gelf.Message{
		Version:  "1.1",
		Host:     h,
		Short:    s,
		Full:     f,
		TimeUnix: float64(time.Now().Unix()),
		Level:    5,
		Extra:    ex,
	}

	return m
}

func unixMilli(t time.Time) int64 {
	return t.Round(time.Millisecond).UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
}

func makeTimestampMilli() int64 {
	return unixMilli(time.Now())
}