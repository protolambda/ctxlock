package ctxlock

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLock(t *testing.T) {
	t.Run("write basic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		require.Len(t, x.write, 1)
		x.Unlock()
		require.Len(t, x.write, 0)
	})
	t.Run("write repeat", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		require.Len(t, x.write, 1)
		x.Unlock()
		require.Len(t, x.write, 0)
		x.Lock()
		require.Len(t, x.write, 1)
		x.Unlock()
		require.Len(t, x.write, 0)
	})
	t.Run("write unlock panic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.Panics(t, func() {
			x.Unlock()
		})
		require.Len(t, x.write, 0)
	})
	t.Run("too many write unlocks", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		require.Len(t, x.write, 1)
		x.Unlock()
		require.Len(t, x.write, 0)
		require.Panics(t, func() {
			x.Unlock()
		})
		require.Len(t, x.write, 0)
	})
	t.Run("read basic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Len(t, x.readers, 0)
	})
	t.Run("read repeat", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Len(t, x.readers, 0)
		x.RLock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Len(t, x.readers, 0)
	})
	t.Run("read unlock panic", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.Panics(t, func() {
			x.RUnlock()
		})
		require.Len(t, x.write, 0)
	})
	t.Run("too many read unlocks", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		require.Len(t, x.write, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Panics(t, func() {
			x.RUnlock()
		})
		require.Len(t, x.write, 0)
	})
	t.Run("read multiple", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RLock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Len(t, x.readers, 0)
	})
	t.Run("ctx lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.NoError(t, x.LockCtx(context.Background()))
		require.Len(t, x.write, 1)
		x.Unlock()
		require.Len(t, x.write, 0)
	})
	t.Run("ctx instant cancel lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // instantly cancel
		err := x.LockCtx(ctx)
		require.NotNil(t, err)
		require.Equal(t, ctx.Err(), err)
	})
	t.Run("ctx later cancel lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		err := x.LockCtx(ctx)
		require.NotNil(t, err)
		require.Equal(t, ctx.Err(), err)
	})
	t.Run("ctx rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.NoError(t, x.RLockCtx(context.Background()))
		require.Len(t, x.write, 1)
		require.Len(t, x.readers, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
		require.Len(t, x.readers, 0)
	})
	t.Run("ctx instant cancel rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // instantly cancel
		err := x.RLockCtx(ctx)
		require.NotNil(t, err)
		require.Equal(t, ctx.Err(), err)
	})
	t.Run("ctx later cancel rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.Lock()
		defer x.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
		defer cancel()
		err := x.RLockCtx(ctx)
		require.NotNil(t, err)
		require.Equal(t, ctx.Err(), err)
	})
	t.Run("ctx multi rlock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.NoError(t, x.RLockCtx(context.Background()))
		require.NoError(t, x.RLockCtx(context.Background()))
		require.NoError(t, x.RLockCtx(context.Background()))
		require.Len(t, x.write, 1)
		x.RUnlock()
		x.RUnlock()
		require.Len(t, x.write, 1)
		x.RUnlock()
		require.Len(t, x.write, 0)
	})
	t.Run("write lock nil ctx", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.Panics(t, func() {
			_ = x.LockCtx(nil) //nolint:all
		})
	})
	t.Run("read lock nil ctx", func(t *testing.T) {
		t.Parallel()
		var x Lock
		require.Panics(t, func() {
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
		require.Panics(t, func() {
			x.RUnlock()
		})
		<-complete
	})
	t.Run("write unlock after read lock", func(t *testing.T) {
		t.Parallel()
		var x Lock
		x.RLock()
		require.Panics(t, func() {
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
