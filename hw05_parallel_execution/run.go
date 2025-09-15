package hw05parallelexecution

import (
	"errors"
	"sync"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if m < 0 {
		return ErrErrorsLimitExceeded
	}
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	totalError := 0
	taskCh := make(chan Task)
	go func() {
		for _, t := range tasks {
			if totalError > m-1 {
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
				if m != 0 && totalError == m {
					break
				}
				if err != nil {
					mu.Lock()
					totalError++
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	if totalError > m-1 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
