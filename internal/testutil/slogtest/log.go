package slogtest

import (
	"testing"

	"golang.org/x/exp/slog"
)

type testWriter struct {
	t testing.TB
}

func (w testWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

// New returns a *slog.Logger that writes to t.Log
func New(t testing.TB) *slog.Logger {
	tw := &testWriter{t: t}
	slh := slog.NewTextHandler(tw)
	return slog.New(slh)
}
