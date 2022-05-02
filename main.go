/*
   Copyright 2016 The Kubernetes Authors.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
       http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"fmt"
	"github.com/microsoft/ApplicationInsights-Go/appinsights"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
	"time"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

func main() {
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
	client.Context().Tags.Cloud().SetRole("hpa-test")
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		pods, err := clientset.CoreV1().Pods("default").List(context.TODO(), v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
		hpas, err := clientset.AutoscalingV2beta2().HorizontalPodAutoscalers("default").List(context.TODO(), v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d hpa in the cluster\n", len(hpas.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		_, err = clientset.CoreV1().Pods("default").Get(context.TODO(), "tools", v1.GetOptions{})
		if errors.IsNotFound(err) {
			fmt.Printf("Pod tools not found in default namespace\n")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
		} else if err != nil {
			panic(err.Error())
		} else {
			fmt.Printf("Found tools pod in default namespace\n")
		}

		for _, hpa := range hpas.Items {
			hpatest, err := clientset.AutoscalingV2beta2().HorizontalPodAutoscalers("default").Get(context.TODO(), hpa.Name, v1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}
			traceMaxHpa := appinsights.NewTraceTelemetry(fmt.Sprintf("The maximum number of replicas for hpa is %d\n", hpatest.Spec.MaxReplicas), appinsights.Information)
			traceMaxHpa.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(traceMaxHpa)
			traceCurrentHpa := appinsights.NewTraceTelemetry(fmt.Sprintf("The current number of replicas for hpa is %d\n", hpatest.Status.CurrentReplicas), appinsights.Information)
			traceCurrentHpa.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(traceCurrentHpa)
			traceHpaReplicasReachMax := appinsights.NewTraceTelemetry(fmt.Sprintf("There are %d pods until reaching maximum autoscaling replicas \n", hpatest.Spec.MaxReplicas-hpatest.Status.CurrentReplicas), appinsights.Information)
			traceHpaReplicasReachMax.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(traceHpaReplicasReachMax)
			fmt.Printf("The maximum number of replicas for hpa is %d\n", hpatest.Spec.MaxReplicas)
			fmt.Printf("The current number of replicas for hpa is %d\n", hpatest.Status.CurrentReplicas)
			metricMaxHpa := appinsights.NewMetricTelemetry("hpaMaxReplicas", float64(hpatest.Spec.MaxReplicas))
			metricMaxHpa.Properties["hpaName"] = hpatest.Spec.ScaleTargetRef.Name
			metricMaxHpa.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(metricMaxHpa)
			metricCurrentHpa := appinsights.NewMetricTelemetry("hpaCurrentReplicas", float64(hpatest.Status.CurrentReplicas))
			metricCurrentHpa.Properties["hpaName"] = hpatest.Spec.ScaleTargetRef.Name
			metricCurrentHpa.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(metricCurrentHpa)
			metricHpaReplicasReachMax := appinsights.NewMetricTelemetry("hpaReplicasUntilReachMax", float64(hpatest.Spec.MaxReplicas-hpatest.Status.CurrentReplicas))
			metricHpaReplicasReachMax.Properties["hpaName"] = hpatest.Spec.ScaleTargetRef.Name
			metricHpaReplicasReachMax.Tags.Cloud().SetRole(hpatest.Spec.ScaleTargetRef.Name)
			client.Track(metricHpaReplicasReachMax)
		}
		//client.TrackMetric("hpaMaxReplicas", float64(hpatest.Spec.MaxReplicas))
		//client.TrackMetric("hpaCurrentReplicas", float64(hpatest.Status.CurrentReplicas))
		time.Sleep(10 * time.Second)
	}
}
