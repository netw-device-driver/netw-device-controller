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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *NetworkNodeReconciler) createNetworkDevice(ctx context.Context, nn *nddv1.NetworkNode) error {

	networkDevice := &nddv1.NetworkDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
			Labels: map[string]string{
				"netwDevice": nn.Name,
			},
		},
		Spec: nddv1.NetworkDeviceSpec{
			Address: nn.Spec.Target.Address,
		},
		Status: nddv1.NetworkDeviceStatus{
			DiscoveryStatus: "Not Ready",
		},
	}
	if err := r.Create(ctx, networkDevice); err != nil {
		if !k8serrors.IsAlreadyExists(err) {
			return err
		}
		ndKey := types.NamespacedName{
			Name:      nn.Name,
			Namespace: nn.Namespace,
		}
		if err := r.Get(ctx, ndKey, networkDevice); err != nil {
			return err
		}
	}
	r.Log.WithValues("NetworkDevice Object", networkDevice).Info("created networkDevice...")

	if err := r.saveNetworkDeviceStatus(ctx, networkDevice); err != nil {
		return err
	}
	return nil
}

func (r *NetworkNodeReconciler) updateNetworkDevice(ctx context.Context, nn *nddv1.NetworkNode) error {
	networkDevice := &nddv1.NetworkDevice{}
	ndKey := types.NamespacedName{
		Name:      nn.Name,
		Namespace: nn.Namespace,
	}
	if err := r.Get(ctx, ndKey, networkDevice); err != nil {
		return err
	}

	networkDevice.Spec = nddv1.NetworkDeviceSpec{
		Address: nn.Spec.Target.Address,
	}
	networkDevice.Status = nddv1.NetworkDeviceStatus{
		DiscoveryStatus: "Not Ready",
	}

	if err := r.Update(ctx, networkDevice); err != nil {
		return err
	}
	r.Log.WithValues("NetworkDevice Object", networkDevice).Info("updated networkDevice...")

	if err := r.saveNetworkDeviceStatus(ctx, networkDevice); err != nil {
		return err
	}

	return nil

}

func (r *NetworkNodeReconciler) deleteNetworkDevice(ctx context.Context, nn *nddv1.NetworkNode) error {
	networkDevice := &nddv1.NetworkDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nn.Name,
			Namespace: nn.Namespace,
			Labels: map[string]string{
				"netwDevice": nn.Name,
			},
		},
	}
	err := r.Delete(ctx, networkDevice)
	if !k8serrors.IsNotFound(err) {
		return err
	}
	r.Log.WithValues("NetworkDevice Object", networkDevice).Info("deleted networkDevice...")
	return nil
}

func (r *NetworkNodeReconciler) saveNetworkDeviceStatus(ctx context.Context, nd *nddv1.NetworkDevice) error {
	t := metav1.Now()
	nd.Status.DeepCopy()
	nd.Status.LastUpdated = &t

	r.Log.Info("Network Node status",
		"status", nd.Status)

	if err := r.Client.Status().Update(ctx, nd); err != nil {
		r.Log.WithValues(nd.Name, nd.Namespace).Error(err, "Failed to update network device status ")
		return err
	}
	return nil
}
