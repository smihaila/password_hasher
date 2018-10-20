package utils

import (
    "sync"
    "time"
)

type WaitGroup sync.WaitGroup

func (wg *WaitGroup) Add(delta int) {
    (*sync.WaitGroup)(wg).Add(delta)
}

func (wg *WaitGroup) Done() {
    (*sync.WaitGroup)(wg).Done()
}

func (wg *WaitGroup) Wait() {
    (*sync.WaitGroup)(wg).Wait()
}

func (wg *WaitGroup) WaitTimeout(timeout time.Duration) bool {
    c := make(chan struct{})
    
    go func() {
        defer close(c)
        wg.Wait()
    }()
    
    select {
    case <-c:
        return false // Completed normally.
    
    case <-time.After(timeout):
        return true // Timed out.
    }
}
