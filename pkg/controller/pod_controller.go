package controller

import (
	"fmt"
	"log"
	"time"

	"github.com/hayeeabdul/kubeidle/pkg/scaler"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type PodController struct {
	clientset *kubernetes.Clientset
	informer  cache.SharedIndexInformer
	scalers   map[string]scaler.Scalable
	active    bool
}

// NewPodController initializes a new PodController.
func NewPodController() (*PodController, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error building in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientset, time.Minute*10, informers.WithNamespace("default"))
	podInformer := informerFactory.Core().V1().Pods().Informer()

	podController := &PodController{
		clientset: clientset,
		informer:  podInformer,
		scalers: map[string]scaler.Scalable{
			"Deployment":  &scaler.DeploymentScaler{},
			"DaemonSet":   &scaler.DaemonSetScaler{},
			"StatefulSet": &scaler.StatefulSetScaler{},
		},
		active: false,
	}

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: podController.handleAdd,
	})

	return podController, nil
}

// SetActiveState sets the controller's active state.
func (pc *PodController) SetActiveState(state bool) {
	pc.active = state
}

// handleAdd handles the logic for scaling down resources when a pod is created.
func (pc *PodController) handleAdd(obj interface{}) {
	if !pc.active {
		log.Println("kubeidle is in stale state. Skipping event handling.")
		return
	}

	pod := obj.(*v1.Pod)
	log.Printf("New Pod Created: %s in Namespace: %s", pod.Name, pod.Namespace)

	for _, ownerRef := range pod.OwnerReferences {
		if scaler, exists := pc.scalers[ownerRef.Kind]; exists {
			err := scaler.ScaleDown(pc.clientset, pod.Namespace, ownerRef.Name)
			if err != nil {
				log.Printf("Error scaling down %s: %v", ownerRef.Kind, err)
			} else {
				log.Printf("Scaled down %s: %s", ownerRef.Kind, ownerRef.Name)
			}
		}
	}
}

// Run starts the pod informer.
func (pc *PodController) Run(stopCh <-chan struct{}) {
	pc.informer.Run(stopCh)
}
