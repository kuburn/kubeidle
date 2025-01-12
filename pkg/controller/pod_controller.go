package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/kuburn/kubeidle/pkg/metrics"
	"github.com/kuburn/kubeidle/pkg/scaler"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			"ReplicaSet":  &scaler.DeploymentScaler{},
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
	if state {
		metrics.ControllerStatus.Set(1)
	} else {
		metrics.ControllerStatus.Set(0)
	}
}

// handleAdd handles the logic for scaling down resources when a pod is created.
func (pc *PodController) handleAdd(obj interface{}) {
	if !pc.active {
		log.Println("kubeidle is inactive. Skipping pod handling.")
		return
	}

	pod := obj.(*v1.Pod)
	log.Printf("New Pod detected: %s in Namespace: %s", pod.Name, pod.Namespace)
	log.Printf("Pod OwnerReferences: %v", pod.OwnerReferences)

	for _, ownerRef := range pod.OwnerReferences {
		if scaler, exists := pc.scalers[ownerRef.Kind]; exists {
			startTime := time.Now()
			err := scaler.ScaleDown(pc.clientset, pod.Namespace, ownerRef.Name)
			duration := time.Since(startTime).Seconds()
			metrics.ReconciliationDuration.WithLabelValues("handle_add").Observe(duration)

			if err != nil {
				metrics.ScaleDownOperationsTotal.WithLabelValues(ownerRef.Kind, pod.Namespace, "failed").Inc()
				log.Printf("Error scaling down %s: %v", ownerRef.Kind, err)
			} else {
				metrics.ScaleDownOperationsTotal.WithLabelValues(ownerRef.Kind, pod.Namespace, "success").Inc()
				metrics.ResourcesCurrentlyScaledDown.WithLabelValues(ownerRef.Kind, pod.Namespace).Inc()
				metrics.LastScaleOperationTimestamp.WithLabelValues(
					ownerRef.Kind,
					pod.Namespace,
					"down",
				).Set(float64(time.Now().Unix()))
				log.Printf("Scaled down %s: %s", ownerRef.Kind, ownerRef.Name)
			}
		}
	}
}

func (pc *PodController) InitialReconcile() error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.ReconciliationDuration.WithLabelValues("initial").Observe(duration)
	}()

	if !pc.active {
		log.Println("kubeidle is inactive. Skipping initial reconciliation.")
		return nil
	}

	log.Println("Starting initial reconciliation of existing Pods...")

	pods, err := pc.clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %v", err)
	}

	for _, pod := range pods.Items {
		log.Printf("Processing existing Pod: %s in Namespace: %s", pod.Name, pod.Namespace)
		log.Printf("Pod OwnerReferences: %v", pod.OwnerReferences)

		for _, ownerRef := range pod.OwnerReferences {
			if scaler, exists := pc.scalers[ownerRef.Kind]; exists {
				err := scaler.ScaleDown(pc.clientset, pod.Namespace, ownerRef.Name)
				if err != nil {
					metrics.ScaleDownOperationsTotal.WithLabelValues(ownerRef.Kind, pod.Namespace, "failed").Inc()
					log.Printf("Error scaling down %s: %v", ownerRef.Kind, err)
				} else {
					metrics.ScaleDownOperationsTotal.WithLabelValues(ownerRef.Kind, pod.Namespace, "success").Inc()
					metrics.ResourcesCurrentlyScaledDown.WithLabelValues(ownerRef.Kind, pod.Namespace).Inc()
					metrics.LastScaleOperationTimestamp.WithLabelValues(
						ownerRef.Kind,
						pod.Namespace,
						"down",
					).Set(float64(time.Now().Unix()))
				}
			}
		}
	}

	return nil
}

func (pc *PodController) Run(stopCh <-chan struct{}) {
	go pc.informer.Run(stopCh) // Start informer in a separate goroutine

	if !cache.WaitForCacheSync(stopCh, pc.informer.HasSynced) {
		log.Println("Failed to sync caches.")
		return
	}

	log.Println("Caches are synced. Controller is ready to process events.")

	for {
		select {
		case <-stopCh:
			log.Println("Stopping PodController.")
			return
		default:
			if pc.active {
				log.Println("Controller is active. Performing initial reconciliation.")
				err := pc.InitialReconcile()
				if err != nil {
					log.Printf("Error during initial reconciliation: %v", err)
				}
				return // Reconciliation should only run once per activation
			}
			time.Sleep(1 * time.Second) // Avoid busy waiting
		}
	}
}
