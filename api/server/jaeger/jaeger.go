package jaeger

import (
	"io"

	"github.com/jaegertracing/jaeger-client-go"
	jaegercfg "github.com/jaegertracing/jaeger-client-go/config"
	jaegerlog "github.com/jaegertracing/jaeger-client-go/log"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-lib/metrics"
)

func InitJaeger() (opentracing.Tracer, io.Closer, error) {
	jaegerConfigInstance := jaegercfg.Configuration{
		ServiceName: "server",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},

		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: "http://jager:16686",
		},
	}

	return jaegerConfigInstance.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
		jaegercfg.Metrics(metrics.NullFactory),
	)
}
