package scaler

import (
	"k8s.io/client-go/kubernetes"
)

// Scalable defines the behavior for scaling Kubernetes resources.
type Scalable interface {
	ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error
}
