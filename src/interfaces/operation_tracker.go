package interfaces

import "time"

type OperationTracker interface {
    BeginOperation(operationName string)
    EndOperation(operationName string)
    WaitForAllOperationsToComplete(timeout time.Duration) bool
}
