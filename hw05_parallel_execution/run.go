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
	mutex := &sync.Mutex{}
	totalError := 0
	countOfTask := len(tasks)
	countOfTaskPerGo := countOfTask / n
	wg := &sync.WaitGroup{}
	if countOfTaskPerGo == 0 {
		countOfTaskPerGo = 1
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if countOfTaskPerGo == 1 && i > countOfTask-1 {
				return
			}
			from := countOfTaskPerGo * i
			to := from + countOfTaskPerGo
			if i == n-1 {
				to += countOfTask % n
			}
			for j := from; j < to; j++ {
				err := tasks[j]()
				mutex.Lock()
				if m != 0 && totalError == m {
					mutex.Unlock()
					break
				}
				if m != 0 && err != nil {
					totalError++
				}
				mutex.Unlock()
			}
		}()
	}
	wg.Wait()
	if totalError > m-1 {
		return ErrErrorsLimitExceeded
	}
	return nil
}
