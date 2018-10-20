package web_api

import (
    "log"
    "fmt"
    "net/http"
    "context"
    "time"
    
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

func (webApi *WebApi) BeginOperation() {
    webApi.pendingOperationsWaitGroup.Add(1)
}

func (webApi *WebApi) EndOperation() {
    webApi.pendingOperationsWaitGroup.Done()
}

func (webApi *WebApi) WaitForAllOperationsToComplete(timeout time.Duration) bool {
    return webApi.pendingOperationsWaitGroup.WaitTimeout(timeout)
}

// -------------------------------------------------------
// WebApi OperationTracker interface implementation end
// -------------------------------------------------------

func (webApi *WebApi) StartServer(statsFilePath string, params ...int) error {
    // TODO: Validate statsFilePath arg.

    const PORT_DEFAULT = 80
    var port int
    {
        paramsLen := len(params)
        if paramsLen > 1 {
            return fmt.Errorf("Only 0 or 1 int args are allowed (%d args provided).", paramsLen)
        }
        
        if 0 == paramsLen {
            port = PORT_DEFAULT
        } else {
            port = params[0]
            if port <= 0 {
                return fmt.Errorf("Invalid port arg: %d.", port)
            }
        }
    }
    
    webApi.statsController.FilePath = statsFilePath
    
    http.HandleFunc("/hash", webApi.onHttpPostHash)
    http.HandleFunc("/stats", webApi.onHttpGetStats)
    http.HandleFunc("/shutdown", webApi.onHttpGetShutdown)
    
    // To serve static files:
    // fs := http.FileServer(http.Dir("static/"))
    // http.Handle("/static/", http.StripPrefix("/static/", fs))
    
    listenAddrAndPort := fmt.Sprintf(":%d", port)
    log.Printf("Web API server listening on %s.", listenAddrAndPort)
    
    //http.ListenAndServe(listenAddrAndPort, nil)
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

func (webApi *WebApi) onHttpPostHash(resp http.ResponseWriter, req *http.Request) {
    defer interfaces.OperationTracker(webApi).EndOperation()
    interfaces.OperationTracker(webApi).BeginOperation()
    
    startTime := time.Now()
    interfaces.Controller(&webApi.hashController).Run(resp, req)
    execDuration := time.Since(startTime)
    
    webApi.statsController.AddHashInfo(float64(execDuration.Nanoseconds()) / 1000.0)
}

func (webApi *WebApi) onHttpGetStats(resp http.ResponseWriter, req *http.Request) {
    defer interfaces.OperationTracker(webApi).EndOperation()
    interfaces.OperationTracker(webApi).BeginOperation()

    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodGet != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }

    interfaces.Controller(&webApi.statsController).Run(resp, req)
}

// The onHttpGetShutdown() is a "Fire and Forget" method. Any WebApi consumer / HTTP client MUST NEITHER expect, NOR
// wait, for a meaningful response.
func (webApi *WebApi) onHttpGetShutdown(resp http.ResponseWriter, req *http.Request) {
    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodGet != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }
    
    for interfaces.OperationTracker(webApi).WaitForAllOperationsToComplete(50 * time.Millisecond) {}
    
    err := webApi.httpServer.Shutdown(context.Background())
    if nil != err {
        log.Fatal(err)
    }
}
