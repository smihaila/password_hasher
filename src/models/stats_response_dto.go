package models

type StatsResponseDTO struct {
    HashCount uint `json:"total"`
    HashAverageResponseTimeUSec float64 `json:"average"`
}
