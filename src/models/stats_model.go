package models

type StatsModel struct {
    HashCount uint `json:"hash_count"`
    HashCumulatedResponseTimeUSec float64 `json:"hash_cumulated_response_time_usec"`
}
