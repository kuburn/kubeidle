package services

import (
	"fmt"
	"time"

	"github.com/kuburn/kubeidle/pkg/database"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// ScaleStateService handles the business logic for managing scale states
type ScaleStateService struct {
	dbConnector database.DBConnector
}

// NewScaleStateService creates a new instance of ScaleStateService
func NewScaleStateService(connector database.DBConnector) (*ScaleStateService, error) {
	if err := connector.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &ScaleStateService{
		dbConnector: connector,
	}, nil
}

// RecordScaleDownState records the state of a scaled-down object
func (s *ScaleStateService) RecordScaleDownState(
	clusterName string,
	obj interface{},
	namespace string,
	replicas int32,
	containers []corev1.Container,
) error {
	var appName, objectType string

	switch v := obj.(type) {
	case *appsv1.Deployment:
		appName = v.Name
		objectType = "Deployment"
	case *appsv1.StatefulSet:
		appName = v.Name
		objectType = "StatefulSet"
	default:
		return fmt.Errorf("unsupported object type")
	}

	// Calculate total resource requests and limits
	var totalCPURequests, totalMemoryRequests, totalCPULimits, totalMemoryLimits string
	for _, container := range containers {
		if container.Resources.Requests != nil {
			cpu := container.Resources.Requests.Cpu()
			memory := container.Resources.Requests.Memory()
			if cpu != nil {
				totalCPURequests = cpu.String()
			}
			if memory != nil {
				totalMemoryRequests = memory.String()
			}
		}
		if container.Resources.Limits != nil {
			cpu := container.Resources.Limits.Cpu()
			memory := container.Resources.Limits.Memory()
			if cpu != nil {
				totalCPULimits = cpu.String()
			}
			if memory != nil {
				totalMemoryLimits = memory.String()
			}
		}
	}

	state := database.ScaleDownState{
		ClusterName:     clusterName,
		ApplicationName: appName,
		Namespace:       namespace,
		ObjectType:      objectType,
		Replicas:        int(replicas),
		CPURequests:     totalCPURequests,
		MemoryRequests:  totalMemoryRequests,
		CPULimits:       totalCPULimits,
		MemoryLimits:    totalMemoryLimits,
		ScaleDownTime:   time.Now().UTC().Format(time.RFC3339),
	}

	return s.dbConnector.InsertScaleDownState(state)
}

// Close closes the database connection
func (s *ScaleStateService) Close() error {
	return s.dbConnector.Close()
}
