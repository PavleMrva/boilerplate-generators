package main

import (
	"fmt"
	"strings"
)

var tracerTemplate string = `package middleware

// TODO: Change <service-package> to the package name of the service

import (
	"context"
	"time"
)

func NewTraceMiddleware(service <service-package>.Service) <service-package>.Service {
	return &traceMiddleware{
		next:        service,
		tracerName: "%s",
	}
}

type traceMiddleware struct {
	next        <service-package>.Service
	tracerName string
}

`

func generateTracerMethodBody(method MethodInfo) string {
	var sb strings.Builder

	params := strings.Split(method.Parameters[0], " ")
	_, name := params[0], params[1]

	ctx := "context.Background()"

	if name == "ctx" {
		ctx = name
	}

	before := fmt.Sprintf(`	// Start a new span
	tracer := otel.Tracer(m.tracerName)

	tracerCtx, span := tracer.Start(%s, "%s")
	defer span.End()
`, ctx, method.Name)

	sb.WriteString(before)

	// Handling for different number of return values
	switch len(method.Returns) {
	case 1:
		// Single return value, could be an error or another type
		if method.Returns[0] == "error" {
			sb.WriteString(fmt.Sprintf("\n\terr := m.next.%s(ctx)\n\n\treturn err", method.Name))
		} else {
			sb.WriteString(fmt.Sprintf("\n\tres := m.next.%s(ctx)\n\n\treturn res", method.Name))
		}
	case 2:
		sb.WriteString(fmt.Sprintf("\n\tres, err := m.next.%s(ctx)\n\n\treturn res, err", method.Name))
	default:
		sb.WriteString(fmt.Sprintf("\n\tm.next.%s(ctx)\n", method.Name))
	}

	sb.WriteString("\n")
	return sb.String()
}

func generateTracerMiddleware(serviceName string, methods []MethodInfo) string {
	initialStr := fmt.Sprintf(tracerTemplate, serviceName)

	for _, method := range methods {
		contentStr := fmt.Sprintf(`func (m *traceMiddleware) %s(`, method.Name)

		for i, param := range method.Parameters {
			p := strings.Split(param, " ")

			if i == len(method.Parameters)-1 {
				contentStr += fmt.Sprintf("%s %s) ", p[1], p[0])
			} else {
				contentStr += fmt.Sprintf("%s %s, ", p[1], p[0])
			}
		}

		if len(method.Returns) > 0 {
			if len(method.Returns) == 1 {
				contentStr += fmt.Sprintf("%s {\n", method.Returns[0])
			} else {
				contentStr += "("

				for i, returns := range method.Returns {
					if i == len(method.Returns)-1 {
						contentStr += fmt.Sprintf("%s) {\n", returns)
					} else {
						contentStr += fmt.Sprintf("%s, ", returns)
					}
				}
			}
		}

		contentStr += generateTracerMethodBody(method)

		contentStr += "}\n\n"
		initialStr += contentStr
	}

	return initialStr
}
