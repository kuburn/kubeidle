# kubeidle

kubeidle is a Kubernetes controller that watches for newly created Pods and scales down the owning resources (such as Deployments, DaemonSets, and StatefulSets) based on a specified time window. This tool is useful for saving resources in non-production environments by automatically scaling down workloads during off-hours.

## Features

	•	Namespace Watcher: Monitors the default namespace for newly created Pods.
	•	Owner Detection: Determines if the owner of the Pod is a Deployment, DaemonSet, or StatefulSet.
	•	Scaling Logic: Scales down the owning resource during a configurable time window.
	•	Configurable Time Window: Specify start and stop times in UTC to control when scaling occurs.
	•	RBAC Support: Ensures necessary permissions for accessing Kubernetes API resources.

## Prerequisites

	•	Kubernetes Cluster: A running Kubernetes cluster.
	•	Helm: Helm 3 or later installed on your system to manage Kubernetes packages.

## Installation

Clone the repository and navigate to the project root:

```bash
git clone https://github.com/yourusername/kubeidle.git
cd kubeidle
```

## Configuration

The primary configuration is managed through environment variables. You can specify the time window in 24-hour format:

	•	START_TIME: Start of the scaling window, e.g., "1800" for 6:00 PM UTC.
	•	STOP_TIME: End of the scaling window, e.g., "0800" for 8:00 AM UTC (next day).

Values are provided in the values.yaml file under the Helm chart and can be overridden during deployment.

**values.yaml**

Adjust any of the default configurations as needed in chart/values.yaml:

```yaml
replicaCount: 1

image:
  repository: your-registry/kubeidle
  tag: latest
  pullPolicy: IfNotPresent

startTime: "1800"
stopTime: "0800"

resources:
  limits:
    cpu: 100m
    memory: 128Mi
  requests:
    cpu: 50m
    memory: 64Mi

rbac:
  create: true  # Set to true if you need RBAC permissions for kubeidle

nodeSelector: {}
tolerations: []
affinity: {}
```

## Deploying kubeidle

To deploy kubeidle using Helm, use the following command in the project root directory:

```bash
helm install kubeidle ./chart --namespace kubeidle --create-namespace
```

This command installs the kubeidle Helm chart with the default settings specified in values.yaml.

If you want to override the start and stop times, you can pass them as arguments during installation:

```bash
helm install kubeidle ./chart --namespace kubeidle --create-namespace --set startTime="1800" --set stopTime="0800"
```

This sets START_TIME to 7:00 PM UTC and STOP_TIME to 7:00 AM UTC for the scaling window.

You can deploy kubeidle directly by referencing the raw URL of this file in your GitHub repository.

```bash
kubectl apply -f https://raw.githubusercontent.com/hayeeabdul/kubeidle/main/manifest/kubeidle.yaml
```

## Updating kubeidle Configuration

To update the time window or other configurations after installation, use Helm’s upgrade command:

```bash
helm upgrade kubeidle ./chart --namespace kubeidle --reuse-values --set startTime="2000" --set stopTime="0600"
```

In case, you want to update the configuration without reinstalling the chart, you can update the ConfigMap directly:

```bash
kubectl delete configmap kubeidle-config -n kubeidle
kubectl create configmap kubeidle-config --from-literal=START_TIME="2000" --from-literal=STOP_TIME="0600" -n kubeidle
```
After that, kubeidle will automatically pick up the new configuration and start scaling the resources accordingly.

## Verifying Deployment

Once deployed, you can check that kubeidle is running with:

```bash
kubectl get pods -n kubeidle
```

You should see a Pod running in the namespace you deployed the chart to (default by default).

## Usage

kubeidle will monitor the default namespace (or the configured namespace if modified) for new Pods. During the specified time window (e.g., 8:00 PM to 8:00 AM UTC), it will scale down the parent Deployment, DaemonSet, or StatefulSet of any newly created Pod it detects. Outside of this window, kubeidle will remain in a dormant state without scaling any resources.

## Uninstalling kubeidle

To remove kubeidle from your cluster, run:

```bash
helm uninstall kubeidle --namespace kubeidle
```

This command deletes all resources associated with the Helm chart.

## Project Structure

The project is organized as follows:

```
kubeidle/
├── cmd/
│   └── kubeidle/
│       └── main.go        # Main entry point
├── pkg/
│   ├── config/            # Time window configuration logic
│   ├── controller/        # Controller logic for monitoring Pods
│   └── scaler/            # Abstraction and scaling logic for various resource types
├── chart/                 # Helm chart for Kubernetes deployment
│   ├── templates/         # Kubernetes resource templates
│   └── values.yaml        # Default configuration values
├── Dockerfile             # Dockerfile to build the kubeidle container image
└── README.md              # Project documentation
```

## Development

	1.	Build the Docker Image:

```bash
docker build -t your-registry/kubeidle .
```
	2.	Push to Container Registry:

```bash
docker push your-registry/kubeidle
```
	3.	Update `values.yaml` to reference your registry and image tag.

## Contributing

Contributions are welcome! Please open an issue or pull request with any improvements or bug fixes.

This README.md covers all aspects of using, configuring, deploying, and maintaining kubeidle with Helm and Docker. Adjustments can be made as needed for further customization.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
