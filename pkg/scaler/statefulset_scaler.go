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

type StatefulSetScaler struct{}

func (s *StatefulSetScaler) ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest StatefulSet object
		statefulset, getErr := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get StatefulSet: %v", getErr)
		}
		log.Println("Successfully retrieved the latest StatefulSet")

		// Create a deep copy to ensure safe modifications
		statefulsetCopy := statefulset.DeepCopy()

		// Update the replica count to zero
		statefulsetCopy.Spec.Replicas = new(int32) // Set replicas to 0
		log.Println("Updating StatefulSet with zero replicas")

		// Apply the update
		_, patchErr := clientset.AppsV1().StatefulSets(namespace).Patch(context.TODO(), statefulsetCopy.Name, types.MergePatchType, []byte(`{"spec":{"replicas":0}}`), metav1.PatchOptions{})
		log.Printf("Current resource version: %s", statefulset.ResourceVersion)
		if patchErr != nil {
			return fmt.Errorf("failed to patch %v", patchErr)
		}
		log.Println("Patched StatefulSet with zero replicas")
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to scale down StatefulSet: %v", retryErr)
	}

	return nil
}
