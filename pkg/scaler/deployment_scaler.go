package scaler

import (
	"context"
	"fmt"

	"github.com/kuburn/kubeidle/pkg/services"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type DeploymentScaler struct {
	client       kubernetes.Interface
	stateService *services.ScaleStateService
	clusterName  string
}

// NewDeploymentScaler creates a new DeploymentScaler
func NewDeploymentScaler(client kubernetes.Interface, stateService *services.ScaleStateService, clusterName string) *DeploymentScaler {
	return &DeploymentScaler{
		client:       client,
		stateService: stateService,
		clusterName:  clusterName,
	}
}

func (d *DeploymentScaler) ScaleDown(namespace string, name string) error {
	// First check if this is a ReplicaSet name, and if so, get the Deployment name
	rs, err := d.client.AppsV1().ReplicaSets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err == nil {
		// Found a ReplicaSet, check if it's owned by a Deployment
		for _, rsOwnerRef := range rs.OwnerReferences {
			if rsOwnerRef.Kind == "Deployment" {
				// Use the Deployment name instead
				name = rsOwnerRef.Name
				break
			}
		}
	}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Fetch the latest Deployment object
		deployment, getErr := d.client.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
		if getErr != nil {
			return fmt.Errorf("failed to get Deployment: %v", getErr)
		}
		logrus.Println("Successfully retrieved the latest Deployment")

		// Create a deep copy to ensure safe modifications
		deploymentCopy := deployment.DeepCopy()

		// Update the replica count to zero
		deploymentCopy.Spec.Replicas = new(int32) // Set replicas to 0
		logrus.Println("Updating Deployment with zero replicas")

		// Apply the update using patch
		_, patchErr := d.client.AppsV1().Deployments(namespace).Patch(context.TODO(), deploymentCopy.Name, types.MergePatchType, []byte(`{"spec":{"replicas":0}}`), metav1.PatchOptions{})
		logrus.Printf("Current resource version: %s", deployment.ResourceVersion)
		if patchErr != nil {
			return fmt.Errorf("failed to patch %v", patchErr)
		}
		logrus.Println("Patched Deployment with zero replicas")

		// After successfully scaling down the deployment
		if err := d.stateService.RecordScaleDownState(
			d.clusterName,
			deployment,
			namespace,
			*deployment.Spec.Replicas,
			deployment.Spec.Template.Spec.Containers,
		); err != nil {
			// Log the error but don't fail the scale down operation
			logrus.Errorf("Failed to record scale down state: %v", err)
		}
		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to scale down Deployment: %v", retryErr)
	}

	return nil
}
