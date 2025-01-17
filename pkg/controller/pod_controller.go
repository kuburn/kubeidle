package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	nsconfig "github.com/kuburn/kubeidle/pkg/config"
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
	informers []cache.SharedIndexInformer
	scalers   map[string]scaler.Scalable
	active    bool
	nsConfig  *nsconfig.NamespaceConfig
}

// NewPodController initializes a new PodController.
func NewPodController(namespaces []string) (*PodController, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error building in-cluster config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating Kubernetes client: %v", err)
	}

	namespaceConfig := nsconfig.NewNamespaceConfig(namespaces)

	podController := &PodController{
		clientset: clientset,
		informers: make([]cache.SharedIndexInformer, 0, len(namespaceConfig.GetNamespaces())),
		scalers: map[string]scaler.Scalable{
			"ReplicaSet":  &scaler.DeploymentScaler{},
			"DaemonSet":   &scaler.DaemonSetScaler{},
			"StatefulSet": &scaler.StatefulSetScaler{},
		},
		active:   false,
		nsConfig: namespaceConfig,
	}

	// Create informers for each namespace
	for _, namespace := range namespaceConfig.GetNamespaces() {
		informer := informers.NewSharedInformerFactoryWithOptions(
			clientset,
			time.Minute*10,
			informers.WithNamespace(namespace),
		).Core().V1().Pods().Informer()

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: podController.handleAdd,
		})

		podController.informers = append(podController.informers, informer)
	}

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

	// Reconcile pods in all configured namespaces
	for _, namespace := range pc.nsConfig.GetNamespaces() {
		pods, err := pc.clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to list pods in namespace %s: %v", namespace, err)
			continue
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
						log.Printf("Scaled down %s: %s", ownerRef.Kind, ownerRef.Name)
					}
				}
			}
		}
	}

	return nil
}

func (pc *PodController) Run(stopCh <-chan struct{}) {
	// Start all informers
	for _, informer := range pc.informers {
		go informer.Run(stopCh)
	}

	// Wait for all caches to sync
	for _, informer := range pc.informers {
		if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
			log.Println("Failed to sync caches.")
			return
		}
	}

	log.Println("Caches are synced. Controller is ready to process events.")

	// Create a ticker to check active state periodically
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stopCh:
			log.Println("Stopping PodController.")
			return
		case <-ticker.C:
			if pc.active {
				log.Println("Controller is active. Performing initial reconciliation.")
				if err := pc.InitialReconcile(); err != nil {
					log.Printf("Error during initial reconciliation: %v", err)
				}
				// Stop the ticker after successful reconciliation
				ticker.Stop()
			}
		}
	}
}
