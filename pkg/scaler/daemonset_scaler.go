package scaler

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type DaemonSetScaler struct{}

func (d *DaemonSetScaler) ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		daemonset, getErr := clientset.AppsV1().DaemonSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Failed to get DaemonSet: %v", getErr)
		}

		daemonset.Spec.Template.Spec.NodeSelector = map[string]string{"kubeidle/disable": "true"} // Example scaling down DaemonSet

		_, updateErr := clientset.AppsV1().DaemonSets(namespace).Update(context.TODO(), daemonset, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("Failed to update DaemonSet: %v", updateErr)
		}

		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to scale down DaemonSet: %v", retryErr)
	}

	return nil
}
