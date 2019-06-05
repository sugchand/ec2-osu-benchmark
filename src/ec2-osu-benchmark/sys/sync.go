
package sys

import (
    "sync"
)

type Sync struct {
    // WaitGroup to keep track of go-routines that are currently running.
    appWaitGroups sync.WaitGroup
}

var appSync = new(Sync)

// Any goroutine invocation must precede with with this function.
// It allows the bookkeeping of currnetly running goroutines in the application.
func (syncObj *Sync)AddRoutineInWaitGroup() {
    syncObj.appWaitGroups.Add(1)
}

// Call when exiting the goroutine after its executing.
// It allows the book-keeping of active gorotuines in the application.
// NEVER INVOKE ExitRoutineInWaitGroup without AddRoutineInWaitGroup
func (syncObj *Sync)ExitRoutineInWaitGroup() {
    syncObj.appWaitGroups.Done()
}

// Function to wait for all the goroutines to complete execution.
// ONLY INVOKED FROM MAIN THREAD AS A LAST STATEMENT.
func (syncObj *Sync)JoinAllRoutines() {
    syncObj.appWaitGroups.Wait()

}

//Function to get the application level syncObj.
func GetAppSyncObj() *Sync{
    return appSync
}