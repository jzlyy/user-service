package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

func TracingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		spanCtx, _ := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(c.Request.Header),
		)

		span := opentracing.GlobalTracer().StartSpan(
			c.FullPath(),
			opentracing.ChildOf(spanCtx),
		)
		defer span.Finish()

		ctx := opentracing.ContextWithSpan(c.Request.Context(), span)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
