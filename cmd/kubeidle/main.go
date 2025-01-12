package main

import (
	"log"
	"os"

	"github.com/kuburn/kubeidle/pkg/config"
	"github.com/kuburn/kubeidle/pkg/controller"
	"github.com/kuburn/kubeidle/pkg/server"
)

func main() {
	startTime := os.Getenv("START_TIME")
	stopTime := os.Getenv("STOP_TIME")
	metricsPort := 9095 // Default metrics port

	if startTime == "" || stopTime == "" {
		log.Fatal("Environment variables START_TIME and STOP_TIME must be set")
	}

	start, stop, err := config.ParseTimeWindow(startTime, stopTime)
	if err != nil {
		log.Fatalf("Error parsing time window: %v", err)
	}

	controller, err := controller.NewPodController()
	if err != nil {
		log.Fatalf("Error creating PodController: %v", err)
	}

	// Start metrics server
	metricsServer := server.NewMetricsServer(metricsPort)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("Error starting metrics server: %v", err)
		}
	}()

	stopCh := make(chan struct{})
	defer close(stopCh)

	go func() {
		for {
			config.WaitForNextActivePeriod(start, stop)
			log.Println("Entering active state for scaling.")
			controller.SetActiveState(true)
			config.WaitForNextStalePeriod(start, stop)
			log.Println("Entering stale state, scaling disabled.")
			controller.SetActiveState(false)
		}
	}()

	log.Println("Starting kubeidle controller")
	controller.Run(stopCh)
}
