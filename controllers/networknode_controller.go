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
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/metal3-io/baremetal-operator/pkg/utils"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nnErrorRetryDelay = time.Minute * 10
)

func init() {
}

// NetworkNodeReconciler reconciles a NetworkNode object
type NetworkNodeReconciler struct {
	client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Namespace string
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;watch;get;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get
// +kubebuilder:rbac:groups="",resources=events,verbs=list;watch;get;patch;create;update;delete
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=devicedrivers,verbs=get;list;watch
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=networkdevices,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=networkdevices/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=networknodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=networknodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=networknodes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NetworkNode object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *NetworkNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (res ctrl.Result, err error) {
	log := r.Log.WithValues("networknode", req.NamespacedName)
	log.Info("reconcile network node...")

	reconcileCounters.With(networkNodeMetricLabels(req)).Inc()
	defer func() {
		if err != nil {
			reconcileErrorCounter.Inc()
		}
	}()

	// get network node information
	nn := &nddv1.NetworkNode{}
	err = r.Client.Get(ctx, req.NamespacedName, nn)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after
			// reconcile request.  Owned objects are automatically
			// garbage collected. For additional cleanup logic use
			// finalizers.  Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, errors.Wrap(err, "could not load network Node data")
	}
	nn.DeepCopy()

	// Add a finalizer to newly created objects.
	if nn.DeletionTimestamp.IsZero() && !networkNodeFinalizer(nn) {
		r.Log.Info(
			"adding finalizer",
			"existingFinalizers", nn.Finalizers,
			"newValue", nddv1.NetworkNodeFinalizer,
		)
		nn.Finalizers = append(nn.Finalizers,
			nddv1.NetworkNodeFinalizer)
		err := r.Update(ctx, nn)
		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to add finalizer")
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// The object is being deleted
	if !nn.DeletionTimestamp.IsZero() && networkNodeFinalizer(nn) {

		if nn.Status.OperationalStatus != nil && *nn.Status.OperationalStatus == nddv1.OperationalStatusUp {
			// only delete the ddriver deployment and networdevice object when operational status is down
			if err = r.deleteNetworkDevice(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					log.Info("failed to delete networkDevice")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete networkDevice"))
				}
			}
			// delete deployment
			if err = r.deleteDeployment(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					log.Info("failed to delete deployment")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete deployment"))
				}
			}
		}
		// remove our finalizer from the list and update it.
		nn.Finalizers = removeString(nn.Finalizers, nddv1.NetworkNodeFinalizer)
		if err := r.Update(context.Background(), nn); err != nil {
			nn.SetOperationalStatus(nddv1.OperationalStatusDown)
			return ctrl.Result{}, errors.Wrap(err,
				fmt.Sprintf("failed to remove finalizer"))
		}
		r.Log.Info("cleanup is complete, removed finalizer",
			"remaining", nn.Finalizers)
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	// Retrieve the Login details from the network node spec and validate
	// the network node details and build the credentials for communicating
	// to the network node.
	creds, err := r.buildAndValidateCredentials(ctx, req, nn)
	if err != nil || creds == nil {
		if nn.Status.OperationalStatus != nil && *nn.Status.OperationalStatus == nddv1.OperationalStatusUp {
			// only delete the ddriver deployment and networdevice object when operational status is down
			if err = r.deleteNetworkDevice(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					log.Info("failed to delete networkDevice")
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					err = r.saveNetworkNodeStatus(ctx, nn)
					if err != nil {
						err = errors.Wrap(err, "failed to update error message")
					}
					//return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete networkDevice"))
				}
			}
			// delete deployment
			if err = r.deleteDeployment(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					log.Info("failed to delete deployment")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete deployment"))
				}
			}
		}
		log.Info("credentials not available or invalid...")
		return r.handleErrorResult(ctx, err, req, nn)
	}

	// retreive device driver information
	c, err := r.buildAndValidateDeviceDriver(ctx, req, nn)
	if err != nil || c == nil {
		if nn.Status.OperationalStatus != nil && *nn.Status.OperationalStatus == nddv1.OperationalStatusUp {
			// only delete the ddriver deployment and networdevice object when operational status is down
			if err = r.deleteNetworkDevice(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					log.Info("failed to delete networkDevice")
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					err = r.saveNetworkNodeStatus(ctx, nn)
					if err != nil {
						err = errors.Wrap(err, "failed to update error message")
					}

					//return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete networkDevice"))
				}
			}
			// delete deployment
			if err = r.deleteDeployment(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					log.Info("failed to delete deployment")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to delete deployment"))
				}
			}
		}
		log.Info("device driver retrieval error...")
		return r.handleErrorResult(ctx, err, req, nn)
	}

	log.WithValues("ddinfo", c).Info("Device Driver information...")

	if nn.Status.OperationalStatus == nil {
		// when the container was never initialized create it
		log.Info("Create Deployment and Network Device when Operational status == nil")
		if err = r.createNetworkDevice(ctx, nn); err != nil {
			log.Info("Failed to create Network Device...")
			return r.handleErrorResult(ctx, err, req, nn)
			//return ctrl.Result{}, errors.Wrap(err,
			//	fmt.Sprintf("failed to create Network Device"))
		}
		if err = r.createDeployment(ctx, nn, c); err != nil {
			log.Info("failed to create deployement")
			return r.handleErrorResult(ctx, err, req, nn)
			//return ctrl.Result{}, errors.Wrap(err,
			//	fmt.Sprintf("failed to create deployemnt"))
		}
	} else {
		if *nn.Status.OperationalStatus != nddv1.OperationalStatusUp {
			// only create the container when operational status is down
			log.Info("Create Deployment and Network Device when Operational status != nil, we should never come here")
			if err = r.createNetworkDevice(ctx, nn); err != nil {
				log.Info("Failed to create Network Device...")
				return r.handleErrorResult(ctx, err, req, nn)
				//return ctrl.Result{}, errors.Wrap(err,
				//	fmt.Sprintf("failed to create Network Device"))
			}
			if err = r.createDeployment(ctx, nn, c); err != nil {
				log.Info("failed to create deployement")
				return r.handleErrorResult(ctx, err, req, nn)
				//return ctrl.Result{}, errors.Wrap(err,
				//	fmt.Sprintf("failed to create deployemnt"))
			}
		} else {
			if !reflect.DeepEqual(c, nn.Status.UsedDeviceDriverSpec.Container) || !reflect.DeepEqual(&nn.Spec, nn.Status.UsedNetworkNodeSpec) {
				// Device Driver spec changes or Network node spec changes
				log.Info("Update Deployment and Network Device after changes to device driver or networkNode spec")
				if err = r.updateNetworkDevice(ctx, nn); err != nil {
					log.Info("Failed to update Network Device...")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to update Network Device"))
				}
				if err = r.updateDeployment(ctx, nn, c); err != nil {
					log.Info("failed to update deployement")
					return r.handleErrorResult(ctx, err, req, nn)
					//return ctrl.Result{}, errors.Wrap(err,
					//	fmt.Sprintf("failed to update deployment"))
				}
			}
		}
	}

	log.Info("deployment created or updated and running...")

	return r.handleOKResult(ctx, nn, c)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetworkNodeReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, option controller.Options) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&nddv1.NetworkNode{}).
		WithOptions(option).
		Watches(
			&source.Kind{Type: &nddv1.DeviceDriver{}},
			handler.EnqueueRequestsFromMapFunc(r.DeviceDriverMapFunc),
		)

	_, err := b.Build(r)
	if err != nil {
		return errors.Wrap(err, "failed setting up with a controller manager")
	}
	return nil
}

// DeviceDriverMapFunc is a handler.ToRequestsFunc to be used to enqeue
// request for reconciliation of FscProxy.
func (r *NetworkNodeReconciler) DeviceDriverMapFunc(o client.Object) []ctrl.Request {
	result := []ctrl.Request{}

	dd, ok := o.(*nddv1.DeviceDriver)
	if !ok {
		panic(fmt.Sprintf("Expected a DeviceDriver but got a %T", o))
	}
	r.Log.WithValues(dd.GetName(), dd.GetNamespace()).Info("DeviceDriver MapFunction")

	selectors := []client.ListOption{
		client.InNamespace(dd.Namespace),
		client.MatchingLabels{},
	}
	nn := &nddv1.NetworkNodeList{}
	if err := r.Client.List(context.TODO(), nn, selectors...); err != nil {
		return result
	}

	for _, n := range nn.Items {
		// only enqueue if the network node device driver kind matches with the device driver label
		for k, v := range dd.GetLabels() {
			if k == "ddriver-kind" {
				if string(*n.Spec.DeviceDriver.Kind) == v {
					name := client.ObjectKey{
						Namespace: n.GetNamespace(),
						Name:      n.GetName(),
					}
					r.Log.WithValues(n.GetName(), n.GetNamespace()).Info("DeviceDriverMapFunc networknode ReQueue")
					result = append(result, ctrl.Request{NamespacedName: name})
				}
			}
		}
	}
	return result
}

func (r *NetworkNodeReconciler) handleOKResult(ctx context.Context, nn *nddv1.NetworkNode, c *corev1.Container) (ctrl.Result, error) {
	saveErr := r.setOKCondition(ctx, nn, c)
	if saveErr != nil {
		return ctrl.Result{Requeue: true}, saveErr
	}
	return ctrl.Result{}, nil
}

func (r *NetworkNodeReconciler) setOKCondition(ctx context.Context, nn *nddv1.NetworkNode, c *corev1.Container) (err error) {
	nn.SetOperationalStatus(nddv1.OperationalStatusUp)
	nn.SetErrorType(nddv1.NoneError)
	nn.SetErrorMessage(stringPtr(""))
	nn.SetUsedDeviceDriverSpec(c)
	nn.SetUsedNetworkNodeSpec(&nn.Spec)
	nn.Status.ErrorCount = intPtr(0)

	err = r.saveNetworkNodeStatus(ctx, nn)
	if err != nil {
		err = errors.Wrap(err, "failed to update OK status")
	}

	return
}

func (r *NetworkNodeReconciler) handleErrorResult(ctx context.Context, err error, req ctrl.Request, nn *nddv1.NetworkNode) (ctrl.Result, error) {

	switch err.(type) {
	// In the event a deployment issue
	// we requeue the network node to handle the issue in the future.
	case *CreateDeploymentError, *UpdateDeploymentError,
		*DeleteDeploymentError:

		deploymentError.Inc()
		saveErr := r.setErrorCondition(ctx, req, nn, nddv1.DeviceDriverError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, req, nn.NewEvent("DeploymentError", err.Error()))

		//return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
		return ctrl.Result{}, errors.Wrap(err, "DeploymentError")

	// In the event a network device issue
	// we requeue the network node to handle the issue in the future.
	case *CreateNetworkDeviceError, *UpdateNetworkDeviceError,
		*GetNetworkDeviceError, *DeleteNetworkDeviceError,
		*SaveNetworkDeviceError:

		networkDeviceError.Inc()
		saveErr := r.setErrorCondition(ctx, req, nn, nddv1.DeviceDriverError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, req, nn.NewEvent("DeviceDriverError", err.Error()))

		//return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
		return ctrl.Result{}, errors.Wrap(err, "DeviceDriverError")
	// In the event a device driver retrieval issue
	// we requeue the network node to handle the issue in the future.
	case *ResolveDeviceDriverRefError:
		deviceDriverError.Inc()
		saveErr := r.setErrorCondition(ctx, req, nn, nddv1.DeviceDriverError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, req, nn.NewEvent("DeviceDriverError", err.Error()))

		//return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
		return ctrl.Result{}, errors.Wrap(err, "DeviceDriverError")

	// In the event a credential secret is defined, but we cannot find it
	// we requeue the network node as we will not know if they create the secret
	// at some point in the future.
	case *ResolveTargetSecretRefError:
		credentialsMissing.Inc()
		saveErr := r.setErrorCondition(ctx, req, nn, nddv1.CredentialError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, req, nn.NewEvent("TargetCredentialError", err.Error()))

		//return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
		return ctrl.Result{}, errors.Wrap(err, "TargetCredentialError")

	// If a Network Node is missing a Target address or secret, or
	// we have found the secret but it is missing the required fields,
	// or the Target address is defined but malformed, we set the
	// network node into an error state but we do not Requeue it
	// as fixing the secret or the Network Node info will trigger
	// the Network Node to be reconciled again
	case *EmptyTargetAddressError, *EmptyTargetSecretError,
		*CredentialsValidationError:
		credentialsInvalid.Inc()
		saveErr := r.setErrorCondition(ctx, req, nn, nddv1.CredentialError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		// Only publish the event if we do not have an error
		// after saving so that we only publish one time.
		r.publishEvent(ctx, req, nn.NewEvent("TargetCredentialError", err.Error()))
		return ctrl.Result{}, nil
	default:
		unhandledError.Inc()
		return ctrl.Result{}, errors.Wrap(err, "An unhandled failure occurred")
	}
}

func (r *NetworkNodeReconciler) setErrorCondition(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode, errType nddv1.ErrorType, message string) (err error) {
	logger := r.Log.WithValues("networknode", req.NamespacedName)

	nn.SetOperationalStatus(nddv1.OperationalStatusDown)
	nn.SetErrorType(errType)
	nn.SetErrorMessage(&message)
	nn.SetUsedDeviceDriverSpec(nil)
	nn.SetUsedNetworkNodeSpec(nil)
	if nn.Status.ErrorCount == nil {
		nn.Status.ErrorCount = new(int)
		*nn.Status.ErrorCount++
	} else {
		*nn.Status.ErrorCount++
	}

	logger.Info(
		"adding error message",
		"message", message,
	)
	err = r.saveNetworkNodeStatus(ctx, nn)
	if err != nil {
		err = errors.Wrap(err, "failed to update error message")
	}

	return
}

func (r *NetworkNodeReconciler) publishEvent(ctx context.Context, req ctrl.Request, event corev1.Event) {
	rLogger := r.Log.WithValues("networknode", req.NamespacedName)
	rLogger.Info("publishing event", "reason", event.Reason, "message", event.Message)
	err := r.Client.Create(ctx, &event)
	if err != nil {
		rLogger.Info("failed to record event, ignoring",
			"reason", event.Reason, "message", event.Message, "error", err)
	}
	return
}

func (r *NetworkNodeReconciler) saveNetworkNodeStatus(ctx context.Context, nn *nddv1.NetworkNode) error {
	t := metav1.Now()
	nn.Status.DeepCopy()
	nn.Status.LastUpdated = &t

	r.Log.Info("Network Node status",
		"status", nn.Status)

	if err := r.Client.Status().Update(ctx, nn); err != nil {
		r.Log.WithValues(nn.Name, nn.Namespace).Error(err, "Failed to update network node ")
		return err
	}
	return nil
}

func networkNodeFinalizer(p *nddv1.NetworkNode) bool {
	return utils.StringInList(p.Finalizers, nddv1.NetworkNodeFinalizer)
}
