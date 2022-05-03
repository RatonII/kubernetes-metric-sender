package main

import (
	"fmt"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	v2beta2 "k8s.io/api/autoscaling/v2beta2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
)

func getAppInsightsClient() appinsights.TelemetryClient {
	appinsights.NewDiagnosticsMessageListener(func(msg string) error {
		fmt.Printf("[%s] %s\n", time.Now().Format(time.UnixDate), msg)
		return nil
	})
	telemetryConfig := appinsights.NewTelemetryConfiguration(os.Getenv("APPINSIGHTS_INSTRUMENTATIONKEY"))
	// Configure how many items can be sent in one call to the data collector:
	telemetryConfig.MaxBatchSize = 8192

	// Configure the maximum delay before sending queued telemetry:
	telemetryConfig.MaxBatchInterval = 2 * time.Second

	client := appinsights.NewTelemetryClientFromConfig(telemetryConfig)
	//client.Context().Tags.Cloud().SetRole("hpa-test")
	return client
}

func getKubernetesClient() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return client
}

func sendHpaMetricTelemetry(hpa *v2beta2.HorizontalPodAutoscaler, metricName string, client appinsights.TelemetryClient, metricValue int32) {
	metricHpa := appinsights.NewMetricTelemetry(metricName, float64(hpa.Spec.MaxReplicas))
	metricHpa.Properties["hpaName"] = hpa.Spec.ScaleTargetRef.Name
	metricHpa.Tags.Cloud().SetRole(hpa.Spec.ScaleTargetRef.Name)
	client.Track(metricHpa)
}

func sendHpaTraceTelemetry(hpa *v2beta2.HorizontalPodAutoscaler, client appinsights.TelemetryClient, customMessage string) {
	traceHpa := appinsights.NewTraceTelemetry(customMessage, appinsights.Information)
	traceHpa.Tags.Cloud().SetRole(hpa.Spec.ScaleTargetRef.Name)
	client.Track(traceHpa)
}
