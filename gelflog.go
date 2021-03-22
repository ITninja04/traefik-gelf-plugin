package plugin_gelflog

import (
	"context"
	"fmt"
	"github.com/kjk/betterguid"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"log"
	"net/http"
	"os"
	"time"
)

var GelfWriter *gelf.UDPWriter
var GelfHostname string
var MWConfig *Config
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
}

// New creates and returns a plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	tLog := &GelfLog{
		Name: name,
		Next: next,
		Config: config,
	}
	if config == nil {
		//log.Fatal("config for Gelf Logger empty")
		return nil, fmt.Errorf("config can not be empty")
	}
	MWConfig = config
	if config.HostnameOverride == "" {
		GelfHostname, _ = os.Hostname()
	}
	GelfWriter, _ = gelf.NewUDPWriter(fmt.Sprintf("%s:%d", config.GelfEndpoint, config.GelfPort))
	return tLog, nil
}


func (h *GelfLog) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	id := betterguid.New()
	if MWConfig.EmitRequestStart {
		req.Header.Set(MWConfig.RequestStartTimeHeader, fmt.Sprint(makeTimestampMilli()))
	}
	if MWConfig.EmitTraceId {
		req.Header.Set(MWConfig.TraceIdHeader, fmt.Sprint(id))
	}
	if GelfWriter != nil {
		var headerMap = map[string]interface{}{}
		for str, val := range req.Header {
			headerMap[str] = val
		}
		headerMap["Host"] = req.Host
		message := wrapMessage(fmt.Sprintf("Request to %s", req.Host), fmt.Sprintf("Request to %s", req.Host), 5, headerMap)
		e := GelfWriter.WriteMessage(message)

		if e != nil {
			log.Println("Received error when sending GELF message:", e.Error())
		}
	}
	h.Next.ServeHTTP(rw, req)
}


func wrapMessage(s string, f string, l int32, ex map[string]interface{}) *gelf.Message {
	/*
		Level is a stanard syslog level
		Facility is deprecated
		Line is deprecated
		File is deprecated
	*/

	m := &gelf.Message{
		Version:  "1.1",
		Host:     GelfHostname,
		Short:    s,
		Full:     f,
		TimeUnix: float64(time.Now().Unix()),
		Level:    l,
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