package scaler

import (
	"context"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type DaemonSetScaler struct{}

func (d *DaemonSetScaler) ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest DaemonSet object
		daemonset, getErr := clientset.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get DaemonSet: %v", getErr)
		}
		log.Println("Successfully retrieved the latest DaemonSet")

		// Create a deep copy to ensure safe modifications
		daemonsetCopy := daemonset.DeepCopy()

		// Update the nodeSelector to prevent scheduling
		log.Println("Updating DaemonSet with nodeSelector")

		// Apply the update using patch
		patchData := []byte(`{"spec":{"template":{"spec":{"nodeSelector":{"kubeidle/disable":"true"}}}}}`)
		_, patchErr := clientset.AppsV1().DaemonSets(namespace).Patch(context.TODO(), daemonsetCopy.Name, types.MergePatchType, patchData, metav1.PatchOptions{})

		if patchErr != nil {
			return fmt.Errorf("failed to patch %v", patchErr)
		}
		log.Println("Patched DaemonSet with nodeSelector")
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to scale down DaemonSet: %v", retryErr)
	}

	return nil
}
