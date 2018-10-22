package web_api

import (
    "log"
    "fmt"
    "net/http"
    "context"
    "time"
    
    "../validation"
    "../interfaces"
    "../controllers"
    "../utils"
)

type WebApi struct {
    httpServer http.Server
    pendingOperationsWaitGroup utils.WaitGroup
    hashController controllers.HashController
    statsController controllers.StatsController
}

// -------------------------------------------------------
// WebApi OperationTracker interface implementation begin
// -------------------------------------------------------

func (webApi *WebApi) BeginOperation(operationName string) {
    if 0 != len(operationName) {
        log.Printf("%v begin.", operationName)
    }
    
    webApi.pendingOperationsWaitGroup.Add(1)
}

func (webApi *WebApi) EndOperation(operationName string) {
    if 0 != len(operationName) {
        log.Printf("%v end.", operationName)
    }
    
    webApi.pendingOperationsWaitGroup.Done()
}

func (webApi *WebApi) WaitForAllOperationsToComplete(timeout time.Duration) bool {
    return webApi.pendingOperationsWaitGroup.WaitTimeout(timeout)
}

// -------------------------------------------------------
// WebApi OperationTracker interface implementation end
// -------------------------------------------------------

func (webApi *WebApi) StartServer(statsConfigFilePath string, params ...int) error {
    if !validation.IsValidOutputFilePath(statsConfigFilePath) {
        return fmt.Errorf("Invalid statsConfigFilePath arg: %v.", statsConfigFilePath)
    }
    
    const PORT_DEFAULT = 80
    var port int
    {
        paramsLen := len(params)
        if paramsLen > 1 {
            return fmt.Errorf("Only 0 or 1 int args are allowed (%v args provided).", paramsLen)
        }
        
        if 0 == paramsLen {
            port = PORT_DEFAULT
        } else {
            port = params[0]
            if !validation.IsValidTcpListenPort(port) {
                return fmt.Errorf("Invalid port arg: %v.", port)
            }
        }
    }
    
    webApi.statsController.Init(statsConfigFilePath)
    
    http.HandleFunc("/hash", webApi.onHttpPostHash)
    http.HandleFunc("/stats", webApi.onHttpGetStats)
    http.HandleFunc("/shutdown", webApi.onHttpGetShutdown)
    
    // To serve static files:
    // fs := http.FileServer(http.Dir("static/"))
    // http.Handle("/static/", http.StripPrefix("/static/", fs))
    
    listenAddrAndPort := fmt.Sprintf(":%v", port)
    log.Printf("Web API server listening on %v.", listenAddrAndPort)
    
    webApi.httpServer = http.Server{
        Addr:           listenAddrAndPort,
        Handler:        nil, // Use DefaultServeMux.
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }
    err := webApi.httpServer.ListenAndServe()
    webApi.httpServer = http.Server{}
    
    if nil != err && http.ErrServerClosed != err {
        log.Fatal(err)
    } else {
        err = nil
        log.Printf("Web API server is now shut down.")
    }
    
    return err
}

// ----------------------------------------
// WebApi private / internal methods begin
// ----------------------------------------

func (webApi *WebApi) onHttpPostHash(resp http.ResponseWriter, req *http.Request) {
    interfaces.OperationTracker(webApi).BeginOperation("WebApi.onHttpPostHash()")
    defer interfaces.OperationTracker(webApi).EndOperation("WebApi.onHttpPostHash()")
    
    startTime := time.Now()
    interfaces.Controller(&webApi.hashController).Run(resp, req)
    execDuration := time.Since(startTime)
    
    webApi.statsController.AddHashInfo(float64(execDuration.Nanoseconds()) / 1000.0)
}

func (webApi *WebApi) onHttpGetStats(resp http.ResponseWriter, req *http.Request) {
    interfaces.OperationTracker(webApi).BeginOperation("WebApi.onHttpGetStats()")
    defer interfaces.OperationTracker(webApi).EndOperation("WebApi.onHttpGetStats()")
    
    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodGet != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }

    interfaces.Controller(&webApi.statsController).Run(resp, req)
}

// The onHttpGetShutdown() is a "Fire and Forget" method. Any WebApi consumer / HTTP client MUST NEITHER expect, NOR
// wait, for any meaningful response.
func (webApi *WebApi) onHttpGetShutdown(resp http.ResponseWriter, req *http.Request) {
    log.Printf("WebApi.onHttpGetShutdown() begin.")
    defer log.Printf("WebApi.onHttpGetShutdown() end.")

    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodGet != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }

    log.Printf("Waiting for any pending operations to complete...")
    for interfaces.OperationTracker(webApi).WaitForAllOperationsToComplete(50 * time.Millisecond) {}
    log.Printf("Done waiting for any pending operations to complete.")
    
    // WARNING: It is of UTMOST IMPORTANCE to perform the WebApi.httpServer.Shutdown() call from within a Goroutine
    // (concurrent) execution context! Otherwise, any calls immediately following after such call will NOT execute and
    // thus, our WebApi.onHttpGetShutdown() method would never end.
    go func() {
        if err := webApi.httpServer.Shutdown(context.Background()); nil != err {
            log.Fatal(err)
        }
    }()
}

// ----------------------------------------
// WebApi private / internal methods end
// ----------------------------------------
