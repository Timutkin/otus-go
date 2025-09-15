package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if m < 0 {
		return ErrErrorsLimitExceeded
	}
	m64 := int64(m)
	var totalError int64
	wg := &sync.WaitGroup{}
	taskCh := make(chan Task)
	go func() {
		for _, t := range tasks {
			if atomic.LoadInt64(&totalError) == m64 {
				break
			}
			taskCh <- t
		}
		close(taskCh)
	}()
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if i > len(tasks) {
				return
			}
			for task := range taskCh {
				err := task()
				if m != 0 && atomic.LoadInt64(&totalError) == m64 {
					break
				}
				if err != nil {
					atomic.AddInt64(&totalError, 1)
				}
			}
		}()
	}
	wg.Wait()
	if atomic.LoadInt64(&totalError) > m64-1 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
