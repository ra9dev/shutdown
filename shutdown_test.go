package shutdown

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShutdownDependencies(t *testing.T) {
	t.Run("flat dependencies shutdown", func(t *testing.T) {
		t.Parallel()

		shutdown := NewGracefulShutdown()

		shutdown.MustAdd("http_server", func(_ context.Context) {
			t.Log("http_server shutdown success")
		})

		shutdown.MustAdd("grpc_server", func(_ context.Context) {
			t.Log("grpc_server shutdown success")
		})

		shutdown.ForceShutdown()
	})

	t.Run("tree dependencies shutdown", func(t *testing.T) {
		t.Parallel()

		shutdown := NewGracefulShutdown()

		dbIsOff := false
		httpServerIsOff := false
		grpcServerIsOff := false

		shutdown.MustAdd("db", func(_ context.Context) {
			dbIsOff = true

			t.Log("db shutdown success")
		})

		shutdown.MustAddDependant("db", "http_server", func(_ context.Context) {
			assert.Truef(t, dbIsOff, "database is off")

			httpServerIsOff = true

			t.Log("http_server shutdown success")
		})

		shutdown.MustAddDependant("db", "grpc_server", func(_ context.Context) {
			assert.Truef(t, dbIsOff, "database is off")

			grpcServerIsOff = true

			t.Log("grpc_server shutdown success")
		})

		shutdown.MustAddDependant("http_server", "cache", func(_ context.Context) {
			assert.Truef(t, httpServerIsOff, "http_server is off")
			assert.Truef(t, grpcServerIsOff, "grpc_server is off")

			t.Log("cache shutdown success")
		})

		shutdown.ForceShutdown()
	})

	t.Run("no dependency root panic", func(t *testing.T) {
		t.Parallel()

		shutdown := NewGracefulShutdown()

		assert.Panics(t, func() {
			shutdown.MustAddDependant("db", "http_server", func(_ context.Context) {
				t.Log("http_server shutdown success")
			})
		})
	})

	t.Run("cyclic dependencies panic", func(t *testing.T) {
		t.Parallel()

		shutdown := NewGracefulShutdown()

		assert.Panics(t, func() {
			shutdown.MustAdd("db", func(ctx context.Context) {
				t.Log("db shutdown success")
			})

			shutdown.MustAddDependant("db", "http_server", func(_ context.Context) {
				t.Log("http_server shutdown success")
			})

			shutdown.MustAddDependant("http_server", "db", func(_ context.Context) {
				t.Log("http_server<->db cycle dependency shutdown success")
			})
		})
	})
}
