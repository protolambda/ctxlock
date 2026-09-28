package ctxlock

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestLock(t *testing.T) {
	t.Run("write basic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		checkChannelLen(t, x.write, 1)
		x.Unlock()
		checkChannelLen(t, x.write, 0)
	})
	t.Run("write repeat", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		checkChannelLen(t, x.write, 1)
		x.Unlock()
		checkChannelLen(t, x.write, 0)
		x.Lock()
		checkChannelLen(t, x.write, 1)
		x.Unlock()
		checkChannelLen(t, x.write, 0)
	})
	t.Run("write unlock panic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkPanics(t, func() {
			x.Unlock()
		})
		checkChannelLen(t, x.write, 0)
	})
	t.Run("too many write unlocks", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		checkChannelLen(t, x.write, 1)
		x.Unlock()
		checkChannelLen(t, x.write, 0)
		checkPanics(t, func() {
			x.Unlock()
		})
		checkChannelLen(t, x.write, 0)
	})
	t.Run("read basic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkChannelLen(t, x.readers, 0)
	})
	t.Run("read repeat", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkChannelLen(t, x.readers, 0)
		x.RLock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkChannelLen(t, x.readers, 0)
	})
	t.Run("read unlock panic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkPanics(t, func() {
			x.RUnlock()
		})
		checkChannelLen(t, x.write, 0)
	})
	t.Run("too many read unlocks", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		checkChannelLen(t, x.write, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkPanics(t, func() {
			x.RUnlock()
		})
		checkChannelLen(t, x.write, 0)
	})
	t.Run("read multiple", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RLock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkChannelLen(t, x.readers, 0)
	})
	t.Run("ctx lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkError(t, x.LockCtx(context.Background()), nil)
		checkChannelLen(t, x.write, 1)
		x.Unlock()
		checkChannelLen(t, x.write, 0)
	})
	t.Run("ctx instant cancel lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // instantly cancel
		err := x.LockCtx(ctx)
		checkError(t, err, ctx.Err())
	})
	t.Run("ctx later cancel lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		err := x.LockCtx(ctx)
		checkError(t, err, ctx.Err())
	})
	t.Run("ctx rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkError(t, x.RLockCtx(context.Background()), nil)
		checkChannelLen(t, x.write, 1)
		checkChannelLen(t, x.readers, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
		checkChannelLen(t, x.readers, 0)
	})
	t.Run("ctx instant cancel rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // instantly cancel
		err := x.RLockCtx(ctx)
		checkError(t, err, ctx.Err())
	})
	t.Run("ctx later cancel rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		err := x.RLockCtx(ctx)
		checkError(t, err, ctx.Err())
	})
	t.Run("ctx multi rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkError(t, x.RLockCtx(context.Background()), nil)
		checkError(t, x.RLockCtx(context.Background()), nil)
		checkError(t, x.RLockCtx(context.Background()), nil)
		checkChannelLen(t, x.write, 1)
		x.RUnlock()
		x.RUnlock()
		checkChannelLen(t, x.write, 1)
		x.RUnlock()
		checkChannelLen(t, x.write, 0)
	})
	t.Run("write lock nil ctx", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkPanics(t, func() {
			_ = x.LockCtx(nil) //nolint:all
		})
	})
	t.Run("read lock nil ctx", func(t *testing.T) {
		t.Parallel()
		var x Lock
		checkPanics(t, func() {
			_ = x.RLockCtx(nil) //nolint:all
		})
	})
	t.Run("read unlock after write lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		complete := make(chan struct{}, 1)
		go func() {
			time.Sleep(time.Second * 4)
			x.Unlock() // to unblock the RUnlock
			complete <- struct{}{}
		}()
		checkPanics(t, func() {
			x.RUnlock()
		})
		<-complete
	})
	t.Run("write unlock after read lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		checkPanics(t, func() {
			x.Unlock()
		})
	})
	t.Run("chaotic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		var v uint64
		var wg sync.WaitGroup
		wg.Add(1000)
		// spawn 1000 routines either writing or reading
		for i := 0; i < 1000; i++ {
			i := i
			if i%2 == 0 {
				// writer sets to illegal value temporarily
				go func() {
					defer wg.Done()
					x.Lock()
					defer x.Unlock()
					v = 1
					time.Sleep(time.Millisecond * 10)
					v = 0
				}()
			} else {
				// readers can run at any time
				go func() {
					defer wg.Done()
					time.Sleep(time.Millisecond * time.Duration(1+(i%100)))
					x.RLock()
					defer x.RUnlock()
					if v != 0 {
						panic("failed, writer state not locked down properly")
					}
				}()
			}
		}
		wg.Wait()
	})
}
