package ctxlock

import "testing"

func checkChannelLen[T any](t *testing.T, ch <-chan T, want int) {
	t.Helper()
	if got := len(ch); got != want {
		t.Fatalf("channel length = %d, want %d", got, want)
	}
}

func checkError(t *testing.T, got, want error) {
	t.Helper()
	if got != want {
		t.Fatalf("error = %v, want %v", got, want)
	}
}

func checkPanics(t *testing.T, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic")
		}
	}()
	fn()
}
