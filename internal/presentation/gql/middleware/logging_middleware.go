package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

type LoggingMiddleware struct{}

func (l *LoggingMiddleware) ExtensionName() string {
	return "LoggingMiddleware"
}

func (l *LoggingMiddleware) Validate(schema graphql.ExecutableSchema) error {
	return nil
}

func (l *LoggingMiddleware) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	tracer := otel.Tracer("graphql-logger")
	ctx, span := tracer.Start(ctx, "GraphQL Request")

	operationContext := graphql.GetOperationContext(ctx)
	operationType := ""
	if operationContext.Operation != nil {
		operationType = string(operationContext.Operation.Operation)
	}

	request := operationContext.RawQuery
	variables := operationContext.Variables

	ip := "unknown"
	httpReq, ok := ctx.Value("httpRequest").(*http.Request)
	if ok {
		ip = getIPAddress(httpReq)
	}

	truncatedQuery := truncateString(request, 200)

	span.SetAttributes(
		attribute.String("graphql.operation.name", operationContext.OperationName),
		attribute.String("graphql.operation.type", operationType),
		attribute.String("graphql.query", truncatedQuery),
		attribute.String("client.ip", ip),
	)

	log.Printf("GraphQL operation started: type=%s, name=%s, client_ip=%s, query=%s, variables=%v",
		operationType, operationContext.OperationName, ip, truncatedQuery, variables)

	resp := next(ctx)

	return func(ctx context.Context) *graphql.Response {
		response := resp(ctx)
		if len(response.Errors) > 0 {
			log.Printf("GraphQL operation failed: %v", response.Errors)
			for _, err := range response.Errors {
				span.RecordError(err)
			}
		} else {
			log.Println("GraphQL operation succeeded")
		}
		span.End()

		return response
	}
}

func NewLoggingMiddleware() graphql.HandlerExtension {
	return &LoggingMiddleware{}
}

func getIPAddress(req *http.Request) string {
	ip := req.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = req.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = strings.Split(req.RemoteAddr, ":")[0]
	}
	return ip
}

func truncateString(str string, num int) string {
	if len(str) > num {
		return str[:num] + "..."
	}
	return str
}
