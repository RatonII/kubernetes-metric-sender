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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	fmt.Println(os.Environ())
	appInsigtsClient := getAppInsightsClient()
	kubernetesClient := getKubernetesClient()
	for {
		// get pods in all the namespaces by omitting namespace
		// Or specify namespace to get pods in particular namespace
		hpas, err := kubernetesClient.AutoscalingV2beta2().HorizontalPodAutoscalers(os.Getenv("KUBE_NAMESPACE")).List(context.TODO(), v1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d hpa in the cluster\n", len(hpas.Items))

		// Examples for error handling:
		// - Use helper functions e.g. errors.IsNotFound()
		// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
		for _, hpa := range hpas.Items {
			hpaDetails, err := kubernetesClient.AutoscalingV2beta2().HorizontalPodAutoscalers(os.Getenv("KUBE_NAMESPACE")).Get(context.TODO(), hpa.Name, v1.GetOptions{})
			if err != nil {
				panic(err.Error())
			}

			sendHpaTraceTelemetry(hpaDetails, appInsigtsClient, fmt.Sprintf("The maximum number of replicas for hpa is %d\n", hpaDetails.Spec.MaxReplicas))
			sendHpaTraceTelemetry(hpaDetails, appInsigtsClient, fmt.Sprintf("The current number of replicas for hpa is %d\n", hpaDetails.Status.CurrentReplicas))
			sendHpaTraceTelemetry(hpaDetails, appInsigtsClient, fmt.Sprintf("There are %d pods until reaching maximum autoscaling replicas \n", hpaDetails.Spec.MaxReplicas-hpaDetails.Status.CurrentReplicas))

			fmt.Printf("The maximum number of replicas for hpa is %d\n", hpaDetails.Spec.MaxReplicas)
			fmt.Printf("The current number of replicas for hpa is %d\n", hpaDetails.Status.CurrentReplicas)

			sendHpaMetricTelemetry(hpaDetails, "hpaMaxReplicas", appInsigtsClient, hpaDetails.Spec.MaxReplicas)
			sendHpaMetricTelemetry(hpaDetails, "hpaCurrentReplicas", appInsigtsClient, hpaDetails.Status.CurrentReplicas)
			sendHpaMetricTelemetry(hpaDetails, "hpaReplicasUntilReachMax", appInsigtsClient, hpaDetails.Spec.MaxReplicas-hpaDetails.Status.CurrentReplicas)
		}
		time.Sleep(10 * time.Second)
	}
}
