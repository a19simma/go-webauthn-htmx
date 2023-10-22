package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func NewLoggerMiddleWare() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		req := c.Request()
		res := c.Response()
		status := res.StatusCode()
		stop := time.Now()
		switch {
		case status >= 100 && status < 400:
			log.Info().
				Str("verb", string(req.Header.Method())).
				Int("status", res.StatusCode()).
				Str("uri", string(req.URI().RequestURI())).
				Str("latency", stop.Sub(start).String()).
				Send()
		default:
			log.Error().
				Str("verb", string(req.Header.Method())).
				Int("status", res.StatusCode()).
				Str("uri", string(req.URI().RequestURI())).
				Str("latency", stop.Sub(start).String()).
				Send()
		}
		return err
	}
}
