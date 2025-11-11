package main

import (
	"emaiorov/load-balancer/config"
	"emaiorov/load-balancer/handlers"
	"log"
	"net/http"
)

func main() {

	var servers []handlers.Server
	appConfig, err := config.Load("config.json")
	if err != nil {
		log.Fatal(err)
	}

	for _, serverConfig := range appConfig.Servers {
		var counter handlers.Counter
		counter.SetLenth(int(serverConfig.Weight))
		servers = append(servers, handlers.Server{
			ServerConfig: serverConfig,
			IsAlive:      true,
			Counter:      counter,
		})
	}

	var handler *handlers.Handler
	var lb handlers.LoadBalancer

	switch appConfig.App.Handler {
	case "LeastConnections":
		lcHandler := handlers.NewLeastConnectionsHandler(servers)
		handler = &lcHandler.Handler
		lb = lcHandler
	default:
		rrHandler := handlers.NewRoundRobinHandler(servers)
		handler = &rrHandler.Handler
		lb = rrHandler
	}

	go handlers.HealthCheck(handler, appConfig.App.HealthCheckSeconds)

	if err := http.ListenAndServe(":"+appConfig.App.Port, lb); err != nil {
		log.Fatal(err)
	}
}
