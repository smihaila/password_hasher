package main

import (
    "./web_api"
)

func main() {
    webApi := web_api.WebApi{}
    webApi.StartServer("./stats.json", 8080)
}
