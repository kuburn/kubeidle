package scaler

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type DeploymentScaler struct{}

func (d *DeploymentScaler) ScaleDown(clientset *kubernetes.Clientset, namespace string, name string) error {
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		deployment, getErr := clientset.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("Failed to get Deployment: %v", getErr)
		}

		deployment.Spec.Replicas = new(int32) // Set replicas to 0

		_, updateErr := clientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if updateErr != nil {
			return fmt.Errorf("Failed to update Deployment: %v", updateErr)
		}

		return nil
	})
	if retryErr != nil {
		return fmt.Errorf("Failed to scale down Deployment: %v", retryErr)
	}

	return nil
}
