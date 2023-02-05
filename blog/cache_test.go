package blog

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func parallelRunning(N int64, fn func()) {
	var wg sync.WaitGroup
	var i int64
	for i = 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	wg.Wait()
}

func TestAccessLimit(t *testing.T) {
	var counter int64
	var limit int64 = 1000
	var count int64

	fn := func() error {
		// to mock a period of time
		time.Sleep(2 * time.Second)
		atomic.AddInt64(&count, 1)
		return nil
	}

	testcases := []struct {
		N, want int64
	}{
		{100, 100},
		{1000, 1000},
		{1200, 1000},
	}

	for _, tc := range testcases {
		parallelRunning(tc.N, func() {
			accessLimit(&counter, limit, "anonymous", fn)
		})
		if count != tc.want {
			t.Fatalf("Parallel=%d, want = %d, but got = %d", tc.N, tc.want, count)
		}
		if counter != 0 {
			t.Fatalf("counter is %d, want initial 0", counter)
		}
		count = 0
	}
}
