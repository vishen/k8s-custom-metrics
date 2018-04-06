package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/util/logs"
	"k8s.io/metrics/pkg/apis/custom_metrics"

	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/cmd/server"
	"github.com/kubernetes-incubator/custom-metrics-apiserver/pkg/provider"
)

const (
	maxHttpRequestCustomMetric = 30
)

var (
	httpRequestCustomMetric = 0
)

type MyCustomMetricsProvider struct{}

func (c *MyCustomMetricsProvider) ListAllMetrics() []provider.MetricInfo {
	metrics := []provider.MetricInfo{
		provider.MetricInfo{
			GroupResource: schema.GroupResource{
				Resource: "pods",
			},
			Namespaced: true,
			Metric:     "http_requests_custom_metric",
		},
		provider.MetricInfo{
			GroupResource: schema.GroupResource{
				Resource: "pods",
			},
			Namespaced: true,
			Metric:     "this-is-my-custom-metric",
		},
	}
	fmt.Printf("ListAllMetrics: %+v\n", metrics)

	return metrics
}

func (c *MyCustomMetricsProvider) GetRootScopedMetricByName(groupResource schema.GroupResource, name string, metricName string) (*custom_metrics.MetricValue, error) {
	fmt.Printf("GetRootScopedMetricByName: groupResource=%+v, name=%s, metricName=%s\n", groupResource, name, metricName)
	return nil, fmt.Errorf("Unable to get scoped metric by name '%s'", metricName)
}

func (c *MyCustomMetricsProvider) GetRootScopedMetricBySelector(groupResource schema.GroupResource, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	fmt.Printf("GetRootScopedMetricBySelector: groupResource=%+v, selector=%+v, metricName=%s\n", groupResource, selector, metricName)
	return nil, fmt.Errorf("Unable to get scoped metric by name '%s'", metricName)
}

func (c *MyCustomMetricsProvider) GetNamespacedMetricByName(groupResource schema.GroupResource, namespace string, name string, metricName string) (*custom_metrics.MetricValue, error) {
	fmt.Printf("GetNamespacedMetricByName: groupResource=%+v, namespace=%s, name=%s, metricName=%s\n", groupResource, namespace, name, metricName)

	if metricName == "http_requests_custom_metric" && name == "sample-metrics-app" && namespace == "default" {
		httpRequestCustomMetric += 1
		if httpRequestCustomMetric > maxHttpRequestCustomMetric {
			httpRequestCustomMetric = 0
		}
		cm := custom_metrics.MetricValue{
			DescribedObject: custom_metrics.ObjectReference{
				APIVersion: groupResource.Group + "/" + k8sruntime.APIVersionInternal,
				Name:       name,
				Namespace:  namespace,
			},
			MetricName: "http_requests_custom_metric",
			Timestamp:  metav1.Time{time.Now()},
			Value:      *resource.NewMilliQuantity(int64(httpRequestCustomMetric*1000.0), resource.DecimalSI),
		}

		fmt.Printf("Found metric: %+v\n", cm)
		return &cm, nil
	}

	return nil, fmt.Errorf("Unable to get scoped metric by name '%s'", metricName)
}

func (c *MyCustomMetricsProvider) GetNamespacedMetricBySelector(groupResource schema.GroupResource, namespace string, selector labels.Selector, metricName string) (*custom_metrics.MetricValueList, error) {
	fmt.Printf("GetNamespacedMetricBySelector: groupResource=%+v, namespace=%s, selector=%+v, metricName=%s\n", groupResource, namespace, selector, metricName)
	return nil, fmt.Errorf("Unable to get scoped metric by name '%s'", metricName)
}

type MyCustomMetricsAdapter struct {
	*server.CustomMetricsAdapterServerOptions
}

func (c MyCustomMetricsAdapter) RunCustomMetricsAdapterServer(stopCh <-chan struct{}) error {
	config, err := c.Config()
	if err != nil {
		return err
	}

	config.GenericConfig.EnableMetrics = true

	/*clientConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("unable to construct lister client config to initialize provider: %v", err)
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(clientConfig)
	if err != nil {
		return fmt.Errorf("unable to construct discovery client for dynamic client: %v", err)
	}

	dynamicMapper, err := dynamicmapper.NewRESTMapper(discoveryClient, apimeta.InterfacesForUnstructured, o.DiscoveryInterval)
	if err != nil {
		return fmt.Errorf("unable to construct dynamic discovery mapper: %v", err)
	}

	clientPool := dynamic.NewClientPool(clientConfig, dynamicMapper, dynamic.LegacyAPIPathResolverFunc)
	if err != nil {
		return fmt.Errorf("unable to construct lister client to initialize provider: %v", err)
	}
	*/

	cmProvider := &MyCustomMetricsProvider{}

	server, err := config.Complete().New("my-custom-metrics-adapter", cmProvider)
	if err != nil {
		return err
	}
	return server.GenericAPIServer.PrepareRun().Run(stopCh)
}

// NewCommandStartPrometheusAdapterServer provides a CLI handler for 'start master' command
func NewCommandStartAdapterServer(out, errOut io.Writer, stopCh <-chan struct{}) *cobra.Command {
	baseOpts := server.NewCustomMetricsAdapterServerOptions(out, errOut)

	customMetricsAdapter := MyCustomMetricsAdapter{
		CustomMetricsAdapterServerOptions: baseOpts,
	}

	cmd := &cobra.Command{
		Short: "Launch the custom metrics API adapter server",
		Long:  "Launch the custom metrics API adapter server",
		RunE: func(c *cobra.Command, args []string) error {
			if err := customMetricsAdapter.Complete(); err != nil {
				return err
			}
			if err := customMetricsAdapter.Validate(args); err != nil {
				return err
			}
			if err := customMetricsAdapter.RunCustomMetricsAdapterServer(stopCh); err != nil {
				return err
			}
			return nil
		},
	}

	flags := cmd.Flags()
	customMetricsAdapter.SecureServing.AddFlags(flags)
	customMetricsAdapter.Authentication.AddFlags(flags)
	customMetricsAdapter.Authorization.AddFlags(flags)
	customMetricsAdapter.Features.AddFlags(flags)

	return cmd
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}

	fmt.Printf("Starting main...\n")

	cmd := NewCommandStartAdapterServer(os.Stdout, os.Stderr, wait.NeverStop)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
