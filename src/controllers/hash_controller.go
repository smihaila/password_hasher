package controllers

import (
    "net/http"
    "time"
    "fmt"

    "../utils"
)

type HashController struct {
}

// ---------------------------------------------------------
// HashController Controller interface implementation begin
// ---------------------------------------------------------

func (hc *HashController) Run(resp http.ResponseWriter, req *http.Request) {
    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodPost != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }

    req.ParseForm()
  
    var password string
    {
        passwordArray := req.Form["password"]
        if 1 != len(passwordArray) {
            http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
            return
        }
        
        // TODO: Perform STRICT input validation i.e. only chars from a predefined list MUST be allowed.
        password = passwordArray[0]
    }
    
    resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
    fmt.Fprintf(resp, utils.HashString(password))
    
    time.Sleep(5 * time.Second)
}

// ---------------------------------------------------------
// HashController Controller interface implementation end
// ---------------------------------------------------------
