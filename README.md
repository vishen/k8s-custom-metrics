# Kubernetes Custom Metrics
This is the most basic implementation for a Kubernetes custom metric server, currently this just serves static metrics.

This has only been tested on Minikube.

## Running on Minikube
```
$ minikube start --extra-config kubelet.EnableCustomMetrics=true
$ minikube addons enable heapster
$ eval (minikube docker-env)
$ make deps
$ make docker_build
$ kubectl apply -f custom_metrics.yaml
```
