package async_utils

import (
	"fmt"
	"runtime"
	"sync"
)

type PoolFunc func()

type easyPool struct {
	limit chan struct{}
	pool  chan PoolFunc
	close PoolFunc
}

func NewPoolFunc(size int, close PoolFunc) *easyPool {
	pool := &easyPool{
		limit: make(chan struct{}, size),
		pool:  make(chan PoolFunc, size),
		close: close,
	}
	go pool.core()
	return pool
}

func (e *easyPool) Send(fn PoolFunc) {
	e.pool <- fn
}

func (e *easyPool) Over() {
	close(e.pool)
}

func (e *easyPool) core() {
	wg := sync.WaitGroup{}
loop:
	for {
		select {
		case fn, over := <-e.pool:
			if !over {
				break loop
			}
			e.limit <- struct{}{}
			wg.Add(1)
			go func(fn PoolFunc) {
				defer func() {
					if err := recover(); err != nil {
						PrintStack()
						fmt.Println("Recover Err: ", err)
					}
				}()

				defer func() {
					<-e.limit
					wg.Done()
				}()
				fn()
			}(fn)
		}
	}

	wg.Wait()
	e.close()
}


func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Printf("GO ==> %s\n", string(buf[:n]))
}