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

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (r *NetworkNodeReconciler) createService(ctx context.Context, nn *nddv1.NetworkNode) error {
	r.Log.WithValues("name", nn.Name).Info("creating service...")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-service-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "nddriver-" + nn.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					Port:       int32(*nn.Spec.GrpcServer.Port),
					TargetPort: intstr.FromInt(*nn.Spec.GrpcServer.Port),
					Protocol:   "TCP",
				},
			},
		},
	}

	err := r.Create(ctx, s)
	if err != nil {
		return err
	}
	r.Log.WithValues("Service Object", s).Info("created service...")

	return nil
}

func (r *NetworkNodeReconciler) deleteService(ctx context.Context, nn *nddv1.NetworkNode) error {
	r.Log.WithValues("name", nn.Name).Info("deleting service...")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-service-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "nddriver-" + nn.Name,
			},
		},
	}
	err := r.Delete(ctx, s)
	if err != nil {
		return err
	}
	r.Log.WithValues("Service Object", s).Info("deleted service...")
	return nil
}

func (r *NetworkNodeReconciler) updateService(ctx context.Context, nn *nddv1.NetworkNode) error {
	r.Log.WithValues("name", nn.Name).Info("updating service...")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-service-" + nn.Name,
			Namespace: r.Namespace,
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": "nddriver-" + nn.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "proxy",
					Port:       int32(*nn.Spec.GrpcServer.Port),
					TargetPort: intstr.FromInt(*nn.Spec.GrpcServer.Port),
					Protocol:   "TCP",
				},
			},
		},
	}

	err := r.Update(ctx, s)
	if err != nil {
		return err
	}
	r.Log.WithValues("Service Object", s).Info("update service...")

	return nil
}
