package trace

import (
	"fmt"
	"net"
	"os"

	"net/http"

	"strings"

	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"

	"strconv"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	log "gitlab.ucloudadmin.com/wu/logrus"
)

var (
	isTracerEnable = false
)

func setTraceEnabled(enable bool) {
	isTracerEnable = enable
	grpc.EnableTracing = enable
}

// InitGlobalOpenTracer 初始化一个全局的Open Tracing的tracer，当前的服务进程会通过 serviceName + hostName来唯一标识
// 	serviceName: 服务的名称
//  hostName: 服务所在的地址
func InitGlobalOpenTracer(serviceName string, hostName string) {
	if cfg := os.Getenv("OPENTRACE_SERVER"); cfg != "" {
		sampleRate := int64(100)
		cfgs := strings.Split(cfg, ",")
		zipkinCollector, err := zipkin.NewHTTPCollector(fmt.Sprintf("http://%s/api/v1/spans", cfgs[0]))
		if err != nil {
			log.WithError(err).Fatal("Fail to init collector")
		}
		if len(cfgs) > 1 {
			sampleRate, err = strconv.ParseInt(cfgs[1], 10, 64)
			if err != nil {
				log.WithError(err).WithField("s", cfgs[1]).Fatal("Second argument must be an integer")
			}
		}
		recorder := zipkin.NewRecorder(zipkinCollector, false, hostName, serviceName)
		openTracer, err := zipkin.NewTracer(recorder,
			zipkin.WithSampler(zipkin.ModuloSampler(uint64(sampleRate))),
			zipkin.ClientServerSameSpan(true))
		if err != nil {
			log.WithError(err).Fatal("Fail to init tracer")
		}
		opentracing.SetGlobalTracer(openTracer)
	}
}

func init() {
	isTracerEnable = true
	grpc.EnableTracing = true

	trace.AuthRequest = AuthRequest
	http.HandleFunc("/control/trace", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "PUT" {
			err := req.ParseForm()
			if err != nil {
				http.Error(w, "only accept form data", 400)
				return
			}
			v := req.PostForm.Get("enable")
			if v == "" {
				http.Error(w, "missing enable field", 400)
				return
			}
			if v == "true" {
				setTraceEnabled(true)
			} else if v == "false" {
				setTraceEnabled(false)
			} else {
				http.Error(w, "invalid enable field: it should be \"true\" or \"false\"", 400)
				return
			}
			fmt.Fprintf(w, "%v", isTracerEnable)
		} else if req.Method == "GET" {
			w.Header().Set("Content-type", "plain/text")
			fmt.Fprintf(w, "%v", isTracerEnable)
		}
	})
}

// TraceInContext will invoke f if there are trace info attached into context
func TraceInContext(ctx context.Context, f func(tr trace.Trace)) {
	if tr, ok := trace.FromContext(ctx); ok {
		f(tr)
	}
}

// IsTraceEnable returns if the trace is enabled
func IsTraceEnable() bool {
	return isTracerEnable
}

// SetTraceEnable enable or disable current trace
func SetTraceEnable(enable bool) {
	isTracerEnable = enable
	grpc.EnableTracing = enable
}

func AuthRequest(req *http.Request) (any, sensitive bool) {
	// RemoteAddr is commonly in the form "IP" or "IP:port".
	// If it is in the form "IP:port", split off the port.
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true, true
	default:
		// if we are in development node, we allow any client to view pages
		return false, false
	}
}
