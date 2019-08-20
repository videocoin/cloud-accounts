package tracer

import (
	"io"

	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

func NewTracer(serviceName string) (io.Closer, error) {
	cfg, err := config.FromEnv()
	if err != nil {
		return nil, err
	}

	cfg.Sampler.Type = jaeger.SamplerTypeConst
	cfg.Sampler.Param = 1
	// set JAEGER_REPORTER_LOG_SPANS to enable
	cfg.Reporter.LogSpans = false

	return cfg.InitGlobalTracer(
		serviceName,
		config.Metrics(metrics.NullFactory),
		config.Logger(jaeger.StdLogger),
	)

}
