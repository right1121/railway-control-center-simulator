package appctx

import (
	"context"
	"net/http"

	"github.com/right1121/railway-control-center-simulator/pkg/logger"
)

type ctxKey struct{}

type Context struct {
	logger *logger.Logger
}

func (c *Context) GetLogger() *logger.Logger {
	return c.logger
}

func NewContext(ctx context.Context, logger *logger.Logger) context.Context {
	c := &Context{
		logger: logger,
	}

	return context.WithValue(ctx, ctxKey{}, c)
}

func FromContext(ctx context.Context) *Context {
	return ctx.Value(ctxKey{}).(*Context)
}

func FromRequest(req *http.Request) *Context {
	return FromContext(req.Context())
}
