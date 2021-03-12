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
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/metal3-io/baremetal-operator/pkg/utils"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	nddv1 "github.com/netw-device-driver/netw-device-controller/api/v1"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	nnErrorRetryDelay = time.Second * 10
)

func init() {
}

// NetworkNodeReconciler reconciles a NetworkNode object
type NetworkNodeReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=list;watch;get;patch;create;update;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get
// +kubebuilder:rbac:groups=ndd.henderiw.be,resources=devicedrivers,verbs=get;list;watch
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

		if nn.Status.OperationalStatus == nddv1.OperationalStatusUp {
			// only delete the container when operational status is down
			if err = r.deleteDeployment(ctx, nn); err != nil {
				if k8serrors.IsNotFound(err) {
					// do nothing
				} else {
					nn.SetOperationalStatus(nddv1.OperationalStatusDown)
					return ctrl.Result{}, errors.Wrap(err,
						fmt.Sprintf("failed to delete container"))
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
		log.Info("credentials not available or invalid...")
		nn.SetOperationalStatus(nddv1.OperationalStatusDown)
		nn.SetErrorType(nddv1.CredentialError)
		nn.SetUsedDeviceDriverSpec(nil)
		nn.SetUsedNetworkNodeSpec(nil)
		if err = r.saveNetworkNodeStatus(ctx, nn); err != nil {
			return ctrl.Result{}, errors.Wrap(err,
				fmt.Sprintf("failed to save networkNode status after credential check"))
		}
		return r.handleErrorResult(ctx, err, req, nn)
	}

	// retreive device driver information
	ddinfo, err := r.buildAndValidateDeviceDriver(ctx, req, nn)
	if err != nil || ddinfo == nil {
		log.Info("device driver retrieval error...")
		nn.SetOperationalStatus(nddv1.OperationalStatusDown)
		nn.SetErrorType(nddv1.DeviceDriverError)
		nn.SetUsedDeviceDriverSpec(nil)
		nn.SetUsedNetworkNodeSpec(nil)
		if err = r.saveNetworkNodeStatus(ctx, nn); err != nil {
			return ctrl.Result{}, errors.Wrap(err,
				fmt.Sprintf("failed to save networkNode status after device driver check"))
		}
		return r.handleErrorResult(ctx, err, req, nn)
	}

	log.WithValues("ddinfo", ddinfo).Info("Device Driver information...")

	if nn.Status.OperationalStatus != nddv1.OperationalStatusUp {
		// only create the container when operational status is down
		if err = r.createDeployment(ctx, nn, ddinfo); err != nil {
			return ctrl.Result{}, errors.Wrap(err,
				fmt.Sprintf("failed to create container"))
		}
	} else {
		if !reflect.DeepEqual(ddinfo, nn.Status.UsedDeviceDriverSpec) || !reflect.DeepEqual(&nn.Spec, nn.Status.UsedNetworkNodeSpec) {
			// device changes
			log.Info("Device Driver spec changes or Network node spec changes")
			if err = r.updateDeployment(ctx, nn, ddinfo); err != nil {
				return ctrl.Result{}, errors.Wrap(err,
					fmt.Sprintf("failed to create container"))
			}
		}
	}

	log.Info("container created or updated and running...")
	nn.SetOperationalStatus(nddv1.OperationalStatusUp)
	nn.SetUsedDeviceDriverSpec(ddinfo)
	nn.SetUsedNetworkNodeSpec(&nn.Spec)
	if err = r.saveNetworkNodeStatus(ctx, nn); err != nil {
		return ctrl.Result{}, errors.Wrap(err,
			fmt.Sprintf("failed to save networkNode status after container creation"))
	}

	return ctrl.Result{}, nil
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
				if string(n.Spec.DeviceDriver.Kind) == v {
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

// Make sure the credentials for the network node are valid
func (r *NetworkNodeReconciler) buildAndValidateCredentials(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode) (creds *Credentials, err error) {
	// Retrieve the secret from Kubernetes for this network node
	credsSecret, err := r.getSecret(ctx, req, nn)
	if err != nil {
		return nil, err
	}

	// Check if address is defined on the network node
	if nn.Spec.Target.Address == "" {
		return nil, &EmptyTargetAddressError{message: "Missing Target connection detail 'Address'"}
	}

	// TODO we could validate the address format

	creds = &Credentials{
		Username: strings.TrimSuffix(string(credsSecret.Data["username"]), "\n"),
		Password: strings.TrimSuffix(string(credsSecret.Data["password"]), "\n"),
	}

	// Verify that the secret contains the expected info.
	err = creds.Validate()
	if err != nil {
		return nil, err
	}

	return creds, nil
}

// Retrieve the secret containing the credentials for talking to the Network Node.
func (r *NetworkNodeReconciler) getSecret(ctx context.Context, request ctrl.Request, nn *nddv1.NetworkNode) (credsSecret *corev1.Secret, err error) {

	if nn.Spec.Target.CredentialsName == "" {
		return nil, &EmptyTargetSecretError{message: "The Target secret reference is empty"}
	}
	secretKey := nn.CredentialsKey()
	credsSecret = &corev1.Secret{}
	err = r.Client.Get(ctx, secretKey, credsSecret)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, &ResolveTargetSecretRefError{message: fmt.Sprintf("The Target secret %s does not exist", secretKey)}
		}
		return nil, err
	}

	return credsSecret, nil
}

func (r *NetworkNodeReconciler) setErrorCondition(ctx context.Context, request ctrl.Request, nn *nddv1.NetworkNode, errType nddv1.ErrorType, message string) (err error) {
	logger := r.Log.WithValues("networknode", request.NamespacedName)

	setErrorMessage(nn, errType, message)

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

// setErrorMessage updates the ErrorMessage in the network Node Status struct
// and increases the ErrorCount
func setErrorMessage(nn *nddv1.NetworkNode, errType nddv1.ErrorType, message string) {
	nn.Status.OperationalStatus = nddv1.OperationalStatusDown
	nn.Status.ErrorType = errType
	nn.Status.ErrorMessage = message
	nn.Status.ErrorCount++
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

func (r *NetworkNodeReconciler) handleErrorResult(ctx context.Context, err error, request ctrl.Request, nn *nddv1.NetworkNode) (ctrl.Result, error) {
	switch err.(type) {
	// In the event a device driver retrieval issue
	// we requeue the network node to handle the issue in the future.
	case *ResolveDeviceDriverRefError:
		deviceDriverError.Inc()
		saveErr := r.setErrorCondition(ctx, request, nn, nddv1.DeviceDriverError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, request, nn.NewEvent("DeviceDriverError", err.Error()))

		return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
	// In the event a credential secret is defined, but we cannot find it
	// we requeue the network node as we will not know if they create the secret
	// at some point in the future.
	case *ResolveTargetSecretRefError:
		credentialsMissing.Inc()
		saveErr := r.setErrorCondition(ctx, request, nn, nddv1.CredentialError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		r.publishEvent(ctx, request, nn.NewEvent("TargetCredentialError", err.Error()))

		return ctrl.Result{Requeue: true, RequeueAfter: nnErrorRetryDelay}, nil
	// If a Network Node is missing a Target address or secret, or
	// we have found the secret but it is missing the required fields,
	// or the Target address is defined but malformed, we set the
	// network node into an error state but we do not Requeue it
	// as fixing the secret or the Network Node info will trigger
	// the Network Node to be reconciled again
	case *EmptyTargetAddressError, *EmptyTargetSecretError,
		*CredentialsValidationError:
		credentialsInvalid.Inc()
		saveErr := r.setErrorCondition(ctx, request, nn, nddv1.CredentialError, err.Error())
		if saveErr != nil {
			return ctrl.Result{Requeue: true}, saveErr
		}
		// Only publish the event if we do not have an error
		// after saving so that we only publish one time.
		r.publishEvent(ctx, request, nn.NewEvent("TargetCredentialError", err.Error()))
		return ctrl.Result{}, nil
	default:
		unhandledCredentialsError.Inc()
		return ctrl.Result{}, errors.Wrap(err, "An unhandled failure occurred with the Target secret")
	}
}

func (r *NetworkNodeReconciler) publishEvent(ctx context.Context, request ctrl.Request, event corev1.Event) {
	rLogger := r.Log.WithValues("networknode", request.NamespacedName)
	rLogger.Info("publishing event", "reason", event.Reason, "message", event.Message)
	err := r.Client.Create(ctx, &event)
	if err != nil {
		rLogger.Info("failed to record event, ignoring",
			"reason", event.Reason, "message", event.Message, "error", err)
	}
	return
}

func networkNodeFinalizer(p *nddv1.NetworkNode) bool {
	return utils.StringInList(p.Finalizers, nddv1.NetworkNodeFinalizer)
}

func (r *NetworkNodeReconciler) buildAndValidateDeviceDriver(ctx context.Context, req ctrl.Request, nn *nddv1.NetworkNode) (ddInfo *nddv1.DeviceDriverSpec, err error) {
	selectors := []client.ListOption{
		client.MatchingLabels{
			"ddriver-kind": string(nn.Spec.DeviceDriver.Kind),
		},
	}
	dds := &nddv1.DeviceDriverList{}
	if err := r.Client.List(ctx, dds, selectors...); err != nil {
		r.Log.WithValues(req.Name, req.Namespace).Error(err, "Failed to get DeviceDrivers")
		return nil, &ResolveDeviceDriverRefError{message: fmt.Sprintf("Failed to get DeviceDrivers")}

	}
	dds.DeepCopy()
	r.Log.WithValues("dds", dds).Info("Device driver Info")

	ddinfo := &nddv1.DeviceDriverSpec{}
	for _, dd := range dds.Items {
		ddinfo = &dd.Spec
	}

	r.Log.WithValues("ddinfo", ddinfo).Info("Device driver Info")

	if ddinfo.Image == "" {
		// apply the default settings
		ddinfo.Image = "henderiw/nats-netwdevicedriver:latest"
		ddinfo.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("20m"),
				corev1.ResourceMemory: resource.MustParse("32Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("250m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		}
	}
	return ddinfo, nil
}

func (r *NetworkNodeReconciler) createContainer(ctx context.Context, nn *nddv1.NetworkNode, ddinfo *nddv1.DeviceDriverSpec) error {
	container := corev1.Container{
		Name:  "nddriver-" + nn.Name,
		Image: ddinfo.Image,
		//Image:           "henderiw/nats-netwdevicedriver:latest",
		ImagePullPolicy: corev1.PullAlways,
		//ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			"--nats-server=nats.default.svc.cluster.local",
			"--topic=" + fmt.Sprintf("srl.%s.*", nn.Name),
			"--protocol=gnmi",
			"--target=172.20.20.5:57400",
			"--username=admin",
			"--password=admin",
			"--skipverify=true",
			"--encoding=JSON_IETF",
		},
		Command: []string{
			"/nats-netwdevicedriver",
		},

		Resources: ddinfo.Resources,
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-" + nn.Name,
			Namespace: "netw-device-controller-system",
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				container,
			},
		},
	}
	err := r.Create(ctx, pod)
	if err != nil {
		return err
	}
	r.Log.WithValues("Pod Object", pod).Info("created pod...")
	return nil
}

func (r *NetworkNodeReconciler) deleteContainer(ctx context.Context, nn *nddv1.NetworkNode) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-" + nn.Name,
			Namespace: "netw-device-controller-system",
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
	}
	err := r.Delete(ctx, pod)
	if err != nil {
		return err
	}
	r.Log.WithValues("Pod Object", pod).Info("deleted pod...")
	return nil
}

func (r *NetworkNodeReconciler) createDeployment(ctx context.Context, nn *nddv1.NetworkNode, ddinfo *nddv1.DeviceDriverSpec) error {
	container := corev1.Container{
		Name:  "nddriver-" + nn.Name,
		Image: ddinfo.Image,
		//Image:           "henderiw/nats-netwdevicedriver:latest",
		ImagePullPolicy: corev1.PullAlways,
		//ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			"--nats-server=nats.default.svc.cluster.local",
			"--topic=" + fmt.Sprintf("srl.%s.*", nn.Name),
			"--protocol=gnmi",
			"--target=172.20.20.5:57400",
			"--username=admin",
			"--password=admin",
			"--skipverify=true",
			"--encoding=JSON_IETF",
		},
		Command: []string{
			"/nats-netwdevicedriver",
		},

		Resources: ddinfo.Resources,
	}

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: "netw-device-controller-system",
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
						container,
					},
				},
			},
		},
	}

	err := r.Create(ctx, deployment)
	if err != nil {
		return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("created deployment...")
	return nil
}

func (r *NetworkNodeReconciler) updateDeployment(ctx context.Context, nn *nddv1.NetworkNode, ddinfo *nddv1.DeviceDriverSpec) error {
	container := corev1.Container{
		Name:  "nddriver-" + nn.Name,
		Image: ddinfo.Image,
		//Image:           "henderiw/nats-netwdevicedriver:latest",
		ImagePullPolicy: corev1.PullAlways,
		//ImagePullPolicy: corev1.PullIfNotPresent,
		Args: []string{
			"--nats-server=nats.default.svc.cluster.local",
			"--topic=" + fmt.Sprintf("srl.%s.*", nn.Name),
			"--protocol=gnmi",
			"--target=172.20.20.5:57400",
			"--username=admin",
			"--password=admin",
			"--skipverify=true",
			"--encoding=JSON_IETF",
		},
		Command: []string{
			"/nats-netwdevicedriver",
		},

		Resources: ddinfo.Resources,
	}

	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: "netw-device-controller-system",
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
						container,
					},
				},
			},
		},
	}

	err := r.Update(ctx, deployment)
	if err != nil {
		return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("updated deployment...")
	return nil
}

func (r *NetworkNodeReconciler) deleteDeployment(ctx context.Context, nn *nddv1.NetworkNode) error {
	deployment := &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nddriver-deployment-" + nn.Name,
			Namespace: "netw-device-controller-system",
			Labels: map[string]string{
				"netwDDriver": "nddriver-" + nn.Name,
			},
		},
	}
	err := r.Delete(ctx, deployment)
	if err != nil {
		return err
	}
	r.Log.WithValues("Deployment Object", deployment).Info("deleted deployment...")
	return nil
}
