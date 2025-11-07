package main

import (
	"emaiorov/load-balancer/config"
	"emaiorov/load-balancer/handlers"
	"log"
	"net/http"
)

func main() {

	var servers []handlers.Server
	appConfig := config.Load()
	for _, serverConfig := range appConfig.Servers {
		servers = append(servers, handlers.Server{
			ServerConfig: serverConfig,
			IsAlive:      true,
		})
	}

	handler := handlers.NewRoundRobinHandler(servers)
	var lb handlers.LoadBalancer = handler

	go handlers.HealthCheck(&handler.Handler, appConfig.App.HealthCheckSeconds)

	if err := http.ListenAndServe(":"+appConfig.App.Port, lb); err != nil {
		log.Fatal(err)
	}
}
