package main

import (
	"fmt"
	"strings"
)

var loggerTemplate string = `package middleware

// TODO: Change <service-package> to the package name of the service

import (
	"context"
	"time"
)

func NewLogMiddleware(service <service-package>.Service) <service-package>.Service {
	return &logMiddleware{
		next:        service,
		serviceName: "%s",
	}
}

type logMiddleware struct {
	next        <service-package>.Service
	serviceName string
}

`

func generateLoggerMethodBody(method MethodInfo) string {
	var sb strings.Builder

	before := "\tlog.WithContext(ctx).WithFields(log.Fields{\n\t\t"
	before += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
`, method.Name)

	for _, param := range method.Parameters {
		p := strings.Split(param, " ")
		_, name := p[0], p[1]

		if name != "ctx" {
			before += fmt.Sprintf("\t\t\"%s\": %s,", name, name)
		}
	}

	before += "\n\t}).Info(\"service-request\")\n\n\t"

	sb.WriteString(before)

	// Handling for different number of return values
	switch len(method.Returns) {
	case 1:
		// Single return value, could be an error or another type
		if method.Returns[0] == "error" {
			sb.WriteString(fmt.Sprintf("err := m.next.%s(ctx)\n\n\t", method.Name))

			rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
			rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"err": 	 err,
	}).Info("service-response")

	return err`, method.Name)

			sb.WriteString(rest)

		} else {
			sb.WriteString(fmt.Sprintf("res := m.next.%s(ctx)\n\n\t", method.Name))

			rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
			rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"res": 	 res,
	}).Info("service-response")

	return res`, method.Name)

			sb.WriteString(rest)
		}
	case 2:
		sb.WriteString(fmt.Sprintf("res, err := m.next.%s(ctx)\n\n\t", method.Name))

		rest := "log.WithContext(ctx).WithFields(log.Fields{\n\t\t"
		rest += fmt.Sprintf(`"service": m.serviceName,
		"method":  "%s",
		"layer":   "service",
		"res": 	 res,
		"err": 	 err,
	}).Info("service-response")

	return res, err`, method.Name)

		sb.WriteString(rest)
	default:
		sb.WriteString(fmt.Sprintf("m.next.%s(ctx)\n", method.Name))
	}

	sb.WriteString("\n")
	return sb.String()
}

func generateLoggerMiddleware(serviceName string, methods []MethodInfo) string {
	initialStr := fmt.Sprintf(loggerTemplate, serviceName)

	for _, method := range methods {
		contentStr := fmt.Sprintf(`func (m *logMiddleware) %s(`, method.Name)

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

		contentStr += generateLoggerMethodBody(method)

		contentStr += "}\n\n"
		initialStr += contentStr
	}

	return initialStr
}
