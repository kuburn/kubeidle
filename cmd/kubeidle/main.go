package main

import (
	"log"
	"os"

	"github.com/kuburn/kubeidle/pkg/config"
	"github.com/kuburn/kubeidle/pkg/controller"
)

func main() {
	startTime := os.Getenv("START_TIME")
	stopTime := os.Getenv("STOP_TIME")

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
