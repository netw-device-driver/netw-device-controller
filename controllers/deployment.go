/*
Copyright 2021 Wim Henderickx.

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

package controllers

import (
	"context"
	"fmt"

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *NetworkNodeReconciler) createDeployment(ctx context.Context, nn *nddv1.NetworkNode, c *corev1.Container) error {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nddriver-" + nn.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nddriver-" + nn.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*c,
					},
				},
			},
		},
	}

	err := r.Create(ctx, deployment)
	if err != nil {
		return &CreateDeploymentError{message: fmt.Sprintf("Failed to create Deployment: %s", err)}
		//return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("created deployment...")
	return nil
}

func (r *NetworkNodeReconciler) updateDeployment(ctx context.Context, nn *nddv1.NetworkNode, c *corev1.Container) error {

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: appv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "nddriver-" + nn.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "nddriver-" + nn.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						*c,
					},
				},
			},
		},
	}

	err := r.Update(ctx, deployment)
	if err != nil {
		return &UpdateDeploymentError{message: fmt.Sprintf("Failed to update Deployment: %s", err)}
		//return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("updated deployment...")
	return nil
}

func (r *NetworkNodeReconciler) deleteDeployment(ctx context.Context, nn *nddv1.NetworkNode) error {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
	}
	err := r.Delete(ctx, deployment)
	if err != nil {
		return &DeleteDeploymentError{message: fmt.Sprintf("Failed to delete Deployment: %s", err)}
		//return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("deleted deployment...")
	return nil
}

func (r *NetworkNodeReconciler) buildAndValidateDeviceDriver(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode) (c *corev1.Container, err error) {
	selectors := []client.ListOption{
		client.MatchingLabels{
			"ddriver-kind": string(*nn.Spec.DeviceDriver.Kind),
		},
	}
	dds := &nddv1.DeviceDriverList{}
	if err := r.Client.List(ctx, dds, selectors...); err != nil {
		r.Log.WithValues(req.Name, req.Namespace).Error(err, "Failed to get DeviceDrivers")
		return nil, &ResolveDeviceDriverRefError{message: fmt.Sprintf("Failed to get DeviceDrivers")}

	}
	dds.DeepCopy()

	for _, dd := range dds.Items {
		c = dd.Spec.Container
	}

	if c == nil {
		r.Log.Info("Using the default device driver configuration")
		// apply the default settings
		envNameSpace := corev1.EnvVar{
			Name: "MY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		}
		envPodIP := corev1.EnvVar{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		}
		c = &corev1.Container{
			Name:            "nddriver-" + nn.Name,
			Image:           "henderiw/netwdevicedriver-gnmi:latest",
			ImagePullPolicy: corev1.PullAlways,
			//ImagePullPolicy: corev1.PullIfNotPresent,
			Args: []string{
				//"--nats-server=nats.default.svc.cluster.local",
				"--cache-server-address=localhost:" + fmt.Sprintf("%d", *nn.Spec.GrpcServer.Port),
				"--device-name=" + fmt.Sprintf("%s", nn.Name),
			},
			Env: []corev1.EnvVar{
				envNameSpace,
				envPodIP,
			},
			Command: []string{
				"/netwdevicedriver-gnmi",
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("20m"),
					corev1.ResourceMemory: resource.MustParse("32Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("250m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
			},
		}
	} else {
		r.Log.Info("Using the specific device driver configuration")
		// update the argument/environment information, since this is specific for the container deployment
		envNameSpace := corev1.EnvVar{
			Name: "MY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		}
		envPodIP := corev1.EnvVar{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "status.podIP",
				},
			},
		}
		c.Args = []string{
			//"--nats-server=nats.default.svc.cluster.local",
			"--cache-server-address=localhost:" + fmt.Sprintf("%d", *nn.Spec.GrpcServer.Port),
			"--device-name=" + fmt.Sprintf("%s", nn.Name),
		}
		c.Env = []corev1.EnvVar{
			envNameSpace,
			envPodIP,
		}
	}
	return c, nil
}
