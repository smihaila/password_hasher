package interfaces

import "net/http"

type Controller interface {
    Run(resp http.ResponseWriter, req *http.Request)
}
