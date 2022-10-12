package shutdown

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGracefulShutdown_Dependencies(t *testing.T) {
	t.Run("flat dependencies shutdown", func(t *testing.T) {
		shutdown := NewGracefulShutdown()

		shutdown.Add("http_server", func(_ context.Context) {
			t.Log("http_server shutdown success")
		})

		shutdown.Add("grpc_server", func(_ context.Context) {
			t.Log("grpc_server shutdown success")
		})

		shutdown.ForceShutdown()
	})

	t.Run("tree dependencies shutdown", func(t *testing.T) {
		shutdown := NewGracefulShutdown()

		dbIsOff := false
		httpServerIsOff := false
		grpcServerIsOff := false

		shutdown.Add("db", func(_ context.Context) {
			dbIsOff = true

			t.Log("db shutdown success")
		})

		shutdown.AddDependant("db", "http_server", func(_ context.Context) {
			assert.Truef(t, dbIsOff, "database is off")

			httpServerIsOff = true

			t.Log("http_server shutdown success")
		})

		shutdown.AddDependant("db", "grpc_server", func(_ context.Context) {
			assert.Truef(t, dbIsOff, "database is off")

			grpcServerIsOff = true

			t.Log("grpc_server shutdown success")
		})

		shutdown.AddDependant("http_server", "cache", func(_ context.Context) {
			assert.Truef(t, httpServerIsOff, "http_server is off")
			assert.Truef(t, grpcServerIsOff, "grpc_server is off")

			t.Log("cache shutdown success")
		})

		shutdown.ForceShutdown()
	})
}
