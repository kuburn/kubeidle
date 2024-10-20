package scaler

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type StatefulSetScaler struct{}

func (s *StatefulSetScaler) ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		statefulset, getErr := clientset.AppsV1().StatefulSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Failed to get StatefulSet: %v", getErr)
		}

		statefulset.Spec.Replicas = new(int32) // Set replicas to 0

		_, updateErr := clientset.AppsV1().StatefulSets(namespace).Update(context.TODO(), statefulset, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("Failed to update StatefulSet: %v", updateErr)
		}

		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to scale down StatefulSet: %v", retryErr)
	}

	return nil
}
