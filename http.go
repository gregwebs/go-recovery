package recovery

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type StackPrintOption string

const (
	StackPrintLines      StackPrintOption = "full"
	StackPrintStructured StackPrintOption = "structured"
	StackPrintNone       StackPrintOption = "none"
)

type SlogHandlerOpts struct {
	StackPrint StackPrintOption
}

func SlogHandler(opts SlogHandlerOpts) func(context.Context, error) {
	return func(ctx context.Context, err error) {
		switch opts.StackPrint {
		case StackPrintStructured:
			slog.ErrorContext(ctx, fmt.Sprintf("%v", err), "full", fmt.Sprintf("%+v", err))
		case StackPrintLines:
			slog.ErrorContext(ctx, fmt.Sprintf("%+v", err))
		case StackPrintNone:
			slog.ErrorContext(ctx, fmt.Sprintf("%v", err))
		default:
			slog.ErrorContext(ctx, fmt.Sprintf("%v", err))
		}
	}
}

type MiddlewareOpts struct {
	ErrorHandler func(context.Context, error)
}

func HTTPMiddleware(opts MiddlewareOpts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			err := Call(func() error {
				next.ServeHTTP(w, r)
				return nil
			})

			if err != nil {
				if errors.Is(err, http.ErrAbortHandler) {
					// we don't recover http.ErrAbortHandler so the response
					// to the client is aborted, this should not be logged
					panic(err)
				}

				if r.Header.Get("Connection") != "Upgrade" {
					w.WriteHeader(http.StatusInternalServerError)
				}

				handler := opts.ErrorHandler
				if handler == nil {
					handler = SlogHandler(SlogHandlerOpts{StackPrint: StackPrintStructured})
				}
				handler(r.Context(), err)
			}
		}

		return http.HandlerFunc(fn)
	}
}
