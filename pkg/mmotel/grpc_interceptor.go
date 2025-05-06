package mmotel

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryTraceInterceptor is the unary interceptor for tracing
func UnaryTraceInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		ctx, span := StartTrace(ctx)
		defer span.End()

		return handler(ctx, req)
	}
}

// StreamTraceInterceptor is the stream interceptor for tracing
func StreamTraceInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx, span := StartTrace(ss.Context())
		defer span.End()

		// wrap server stream to inject new context
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}
		return handler(srv, wrapped)
	}
}

// wrappedServerStream wraps the grpc.ServerStream to override context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
