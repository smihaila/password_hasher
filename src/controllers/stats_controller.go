package controllers

import (
    "log"
    "fmt"
    "net/http"
    "sync"
    "os"
    "io/ioutil"
    "encoding/json"
    "math"
    
    "../validation"
    "../models"
)

type StatsController struct {
    configFilePath string

    mutex sync.Mutex
    hashCount uint
    hashCumulatedResponseTimeUSec float64
}

func (sc *StatsController) Init(configFilePath string) error {
    if !validation.IsValidOutputFilePath(configFilePath) {
        return fmt.Errorf("Invalid configFilePath arg: %v.", configFilePath)
    }
    
    sc.configFilePath = configFilePath
    return nil
}

func (sc *StatsController) ConfigFilePath() string {
    return sc.configFilePath
}

// ----------------------------------------------------------
// StatsController Controller interface implementation begin
// ----------------------------------------------------------

func (sc *StatsController) Run(resp http.ResponseWriter, req *http.Request) {
    // The resp arg is assumed to be always non-nil.
    if nil == req || http.MethodGet != req.Method {
        http.Error(resp, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
        return
    }
    
    var statsResponseDTO models.StatsResponseDTO
    func() {
        sc.mutex.Lock()
        defer sc.mutex.Unlock()
        
        statsResponseDTO = models.StatsResponseDTO{
            HashCount: sc.hashCount,
            HashAverageResponseTimeUSec: sc.hashAverageResponseTimeUSec(),
        }
    }()
    
    resp.Header().Set("Content-Type", "application/json")
    json.NewEncoder(resp).Encode(statsResponseDTO)
}

// ----------------------------------------------------------
// StatsController Controller interface implementation end
// ----------------------------------------------------------

func (sc *StatsController) AddHashInfo(hashResponseTimeUSec float64) error {
    sc.mutex.Lock()
    defer sc.mutex.Unlock()
    
    err := sc.load()
    if nil != err {
        log.Fatal(err)
        return err
    }
    
    sc.hashCount++
    sc.hashCumulatedResponseTimeUSec += hashResponseTimeUSec
    
    err = sc.save()
    if nil != err {
        log.Fatal(err)
        return err
    }
    
    return nil
}

// -------------------------------------------------
// StatsController private / internal methods begin
// -------------------------------------------------

// WARNING: The StatsController.load() private method MUST ALWAYS be called within an acquired StatsController.mutex
// context!
func (sc *StatsController) load() error {
    if _, err := os.Stat(sc.configFilePath); os.IsNotExist(err) {
        sc.hashCount = 0
        sc.hashCumulatedResponseTimeUSec = 0
        return nil
    }
    
    jsonBytes, err := ioutil.ReadFile(sc.configFilePath)
    if nil != err {
        log.Fatal(err)
        return err
    }
    
    var statsModel models.StatsModel
    err = json.Unmarshal(jsonBytes, &statsModel)
    if nil != err {
        log.Fatal(err)
        return err
    }
    
    sc.hashCount = statsModel.HashCount
    sc.hashCumulatedResponseTimeUSec = statsModel.HashCumulatedResponseTimeUSec
    return nil
}

// WARNING: The StatsController.save() private method MUST ALWAYS be called within an acquired StatsController.mutex
// context!
func (sc *StatsController) save() error {
    statsModel := models.StatsModel {
        HashCount: sc.hashCount,
        HashCumulatedResponseTimeUSec: sc.hashCumulatedResponseTimeUSec,
    }
    jsonBytes, err := json.Marshal(statsModel)
    if nil != err {
        log.Fatal(err)
        return err
    }
    
    err = ioutil.WriteFile(sc.configFilePath, jsonBytes, os.ModePerm)
    if nil != err {
        log.Fatal(err)
        return err
    }

    return nil
}

// WARNING: The StatsController.hashAverageResponseTimeUSec() private method MUST ALWAYS be called within an acquired
// StatsController.mutex context!
func (sc *StatsController) hashAverageResponseTimeUSec() float64 {
    if err := sc.load(); nil != err {
        log.Fatal(err)
        return -3
    }

    if 0 == sc.hashCount {
        return -1
    }
    
    avgRespTime := sc.hashCumulatedResponseTimeUSec / float64(sc.hashCount)
    if math.IsInf(avgRespTime, 0) {
        return -2
    }
    
    return avgRespTime
}

// -------------------------------------------------
// StatsController private / internal methods end
// -------------------------------------------------

