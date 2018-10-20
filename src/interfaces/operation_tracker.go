package interfaces

import "time"

type OperationTracker interface {
    BeginOperation()
    EndOperation()
    WaitForAllOperationsToComplete(timeout time.Duration) bool
}
