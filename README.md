# Kubernetes Custom Metrics and Horizontal Pod Autoscaler
This is the most basic implementation for a Kubernetes custom metric server, currently this just serves a static metric `http_requests_custom_metric` that increments by one eveytime it is asked about the metric.

In this example everytime Kubernetes gets the value of `http_requests_custom_metric` it increments by one, which will be reflected in the HPA, and since this is the value the HPA uses to autoscale it is effectively autoscaling itself, so you can `kubectl get (pods|hpa) -w`. Currently every 30 requests teh counter will be reset to 0.

Currently this is completely hardcoded with if statements in the `GetNamespacedMetricByName` method (which is what the Kubernetes uses to determine a custom metric value, I think?). It should be trivial to implement a more complex custom metrics api by implementing the following methods on on `MyCustomMetricsProvider`:
```
- ListAllMetrics
- GetRootScopedMetricByName
- GetRootScopedMetricBySelector
- GetNamespacedMetricByName
- GetNamespacedMetricBySelector
```

and then using the metric name in the HPA `metricName: http_requests_custom_metric`.


NOTE: This has only been tested on Minikube.

## Running on Minikube

### Starting Minikube
```
$ minikube start --extra-config kubelet.EnableCustomMetrics=true
$ minikube addons enable heapster
```

### Running Kubernetes custom metric API server
```
$ glide install -v
$ eval (minikube docker-env)
$ make docker_build
$ docker build -t k8s-custom-metrics .
$ kubectl apply -f k8s_custom_metrics.yaml
```

### Running sample HPA application
```
$ eval (minikube docker-env)
$ docker build -t k8s-custom-metrics-sample-app sample_app/
$ kubectl apply -f k8s_sample_app.yaml
```

### Checking that it is working
```
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1
$ kubectl get --raw /apis/custom.metrics.k8s.io/v1beta1/namespaces/default/pod/sample-metrics-app/http_requests_custom_metric | jq
```

### Watching it autoscale
Run the following in two seperate terminals to see Kubernetes HPA in action
```
$ kubectl get pods -w
```
```
$ kubectl get hpa -w
```

## Resources
```
- https://github.com/DirectXMan12/k8s-prometheus-adapter
- https://github.com/luxas/kubeadm-workshop
- https://blog.jetstack.io/blog/resource-and-custom-metrics-hpa-v2/
```
