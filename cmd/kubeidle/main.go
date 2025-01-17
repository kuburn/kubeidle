package main

import (
	"log"
	"os"
	"strings"

	"github.com/kuburn/kubeidle/pkg/config"
	"github.com/kuburn/kubeidle/pkg/controller"
	"github.com/kuburn/kubeidle/pkg/server"
)

func main() {
	log.Println("Starting kubeidle controller...")
	
	startTime := os.Getenv("START_TIME")
	stopTime := os.Getenv("STOP_TIME")
	metricsPort := 9095
	
	log.Printf("Configuration - Start Time: %s, Stop Time: %s", startTime, stopTime)
	
	// Get namespaces from environment variable
	namespacesStr := os.Getenv("NAMESPACES")
	var namespaces []string
	if namespacesStr != "" {
		namespaces = strings.Split(namespacesStr, ",")
		log.Printf("Configured namespaces: %v", namespaces)
	} else {
		log.Println("No namespaces specified, will use default namespace")
	}

	if startTime == "" || stopTime == "" {
		log.Fatal("Environment variables START_TIME and STOP_TIME must be set")
	}

	log.Println("Parsing time window...")
	start, stop, err := config.ParseTimeWindow(startTime, stopTime)
	if err != nil {
		log.Fatalf("Error parsing time window: %v", err)
	}
	log.Printf("Time window parsed successfully: start=%v, stop=%v", start, stop)

	log.Println("Creating PodController...")
	controller, err := controller.NewPodController(namespaces)
	if err != nil {
		log.Fatalf("Error creating PodController: %v", err)
	}
	log.Println("PodController created successfully")

	// Start metrics server
	log.Printf("Starting metrics server on port %d...", metricsPort)
	metricsServer := server.NewMetricsServer(metricsPort)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Printf("Error starting metrics server: %v", err)
		}
	}()

	stopCh := make(chan struct{})
	defer close(stopCh)

	log.Println("Starting controller scheduling goroutine...")
	go func() {
		for {
			log.Println("Waiting for next active period...")
			config.WaitForNextActivePeriod(start, stop)
			log.Println("Entering active state for scaling.")
			controller.SetActiveState(true)
			log.Println("Waiting for next stale period...")
			config.WaitForNextStalePeriod(start, stop)
			log.Println("Entering stale state.")
			controller.SetActiveState(false)
		}
	}()

	log.Println("Starting main controller loop...")
	controller.Run(stopCh)
}
