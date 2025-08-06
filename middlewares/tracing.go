package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		spanCtx, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		span := opentracing.StartSpan(
			c.FullPath(),
			ext.RPCServerOption(spanCtx),
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			opentracing.Tag{Key: "http.method", Value: c.Request.Method},
			opentracing.Tag{Key: "http.url", Value: c.Request.URL.String()},
		)
		defer span.Finish()

		// 注入到上下文
		ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		// 记录状态码
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
		if c.Writer.Status() >= http.StatusInternalServerError {
			ext.Error.Set(span, true)
		}
	}
}
