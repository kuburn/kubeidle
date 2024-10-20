# KubeIdle

KubeIdle is a Kubernetes controller designed to automatically scale down idle pods to save resources and costs in your Kubernetes cluster.

## Features

- Automatically detects idle pods based on configurable metrics
- Scales down idle pods to zero replicas
- Scales up pods when traffic resumes
- Configurable idle detection parameters

## Installation

To install KubeIdle in your Kubernetes cluster:

bash
kubectl apply -f https://raw.githubusercontent.com/hayeeabdul/kubeidle/main/deploy/kubeidle.yaml


## Configuration

KubeIdle can be configured via a ConfigMap. Here's an example configuration:

yaml
apiVersion: v1
kind: ConfigMap
metadata:
name: kubeidle-config
namespace: kube-system
data:
config.yaml: |
idleThreshold: 5m
checkInterval: 1m
excludedNamespaces:
kube-system
monitoring



## Usage

Once installed, KubeIdle will automatically monitor your pods and scale down those that are idle. No additional action is required.

## Development

To set up the development environment:

1. Clone the repository:
   ```
   git clone https://github.com/hayeeabdul/kubeidle.git
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

3. Run the controller locally:
   ```
   go run cmd/kubeidle/main.go
   ```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.