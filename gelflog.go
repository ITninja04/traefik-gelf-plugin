package traefik_gelf_plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kjk/betterguid"
	"gopkg.in/Graylog2/go-gelf.v2/gelf"
	"log"
	"net/http"
	"os"
	"reflect"
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
	Debug bool `json:"debug,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{
		EmitTraceId: true,
		TraceIdHeader: "X-TraceId",
		EmitRequestStart: true,
		RequestStartTimeHeader: "X-Request-Start",
		Debug: false,
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
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	log.Println("DEV MODE ENABLED!")
	tLog := &GelfLog{
		Name: name,
		Next: next,
		Config: config,
	}
	s, err := json.MarshalIndent(ctx, "", "\t")
	log.Println("Plugin Context")
	if err != nil {
		log.Println("Error printing context", err)
	}else {
		log.Println(string(s))
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
	if config.Debug {
		log.Println(fmt.Sprintf("Debug Logging Enabled"), config)
		var configMap = map[string]interface{}{}
		v := reflect.ValueOf(config)
		typeOfS := v.Elem().Type()
		for i := 0; i< v.Elem().NumField(); i++ {
			configMap[typeOfS.Field(i).Name] = v.Elem().Field(i).Interface()
		}
		tLog.GelfWriter.WriteMessage(wrapMessage("Logger Initialized", "Logger successfully initialized", tLog.GelfHostname, configMap))
		log.Println("Sent message to GELF Endpoint")
	}
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
		err := h.GelfWriter.WriteMessage(message)
		if h.Config.Debug {
			log.Println("Sent message to GELF Endpoint", err)
			log.Println("Request Context")
			s, err := json.MarshalIndent(req.Context(), "", "\t")
			s1, err1 := json.MarshalIndent(context.Background(), "", "\t")

			if err != nil {
				log.Println("Error printing context", err)
			}else {
				log.Println(string(s))
			}
			log.Println("Background Context")
			if err1 != nil {
				log.Println("Error printing background context", err)
			}else {
				log.Println(string(s1))
			}
		}
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