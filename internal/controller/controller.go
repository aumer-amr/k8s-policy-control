package controller

import (
	"context"
	"os"
	"strconv"

	"github.com/aumer-amr/k8s-policy-control/internal/policy"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	controllerLog              = ctrl.Log.WithName("controller")
	controllerTriggerUid       = "policy-control.aumer.io/controller-trigger-uid"
	controllerTriggerGroup     = "policy-control.aumer.io/controller-trigger-group"
	controllerTriggerKind      = "policy-control.aumer.io/controller-trigger-kind"
	controllerTriggerVersion   = "policy-control.aumer.io/controller-trigger-version"
	controllerTriggerNamespace = "policy-control.aumer.io/controller-trigger-namespace"
)

type ReconcilerHandler struct {
	Client     client.Client
	PolicyType int
	Manager    ctrl.Manager
	Controller controller.Controller
}

func New(mgr ctrl.Manager, policyType int) *ReconcilerHandler {
	controller := &ReconcilerHandler{
		PolicyType: policyType,
		Manager:    mgr,
		Client:     mgr.GetClient(),
	}
	controller.SetupWithManager(mgr)
	return controller
}

func (r *ReconcilerHandler) SetupWithManager(mgr ctrl.Manager) {
	c, err := controller.New("controller-policy-"+strconv.Itoa(r.PolicyType), mgr, controller.Options{
		Reconciler: &ReconcilerHandler{
			Client:     mgr.GetClient(),
			PolicyType: r.PolicyType,
			Manager:    mgr,
		},
	})
	if err != nil {
		controllerLog.Error(err, "unable to set up individual controller", "PolicyType", r.PolicyType)
		os.Exit(1)
	}

	controllerLog.Info("Setting up controller", "PolicyType", r.PolicyType)

	r.Controller = c

	if r.PolicyType == policy.PolicyTypeIngress {
		r.WatchResource(mgr, &networkingv1.Ingress{}, controllerLog)
	} else if r.PolicyType == policy.PolicyTypePod {
		r.WatchResource(mgr, &corev1.Pod{}, controllerLog)
	} else if r.PolicyType == policy.PolicyTypeUnknown {
		controllerLog.Info("PolicyType is unknown, not watching any resources")
	}
}

func (r *ReconcilerHandler) WatchResource(mgr ctrl.Manager, resourceType client.Object, log logr.Logger) {
	if err := r.Controller.Watch(source.Kind(mgr.GetCache(), resourceType), &handler.EnqueueRequestForObject{}); err != nil {
		log.Error(err, "unable to watch resource")
		os.Exit(1)
	}

	if err := r.Controller.Watch(source.Kind(mgr.GetCache(), resourceType),
		handler.EnqueueRequestForOwner(mgr.GetScheme(), mgr.GetRESTMapper(), resourceType, handler.OnlyControllerOwner())); err != nil {
		log.Error(err, "unable to watch Pods")
		os.Exit(1)
	}
}

func (r *ReconcilerHandler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog.Info("Reconciling", "request", req)

	if r.PolicyType == policy.PolicyTypeIngress {
		return r.reconcileIngress(ctx, req, controllerLog)
	} else if r.PolicyType == policy.PolicyTypePod {
		//return r.reconcilePod(ctx, req, controllerLog)
		return ctrl.Result{}, nil
	} else {
		return ctrl.Result{}, nil
	}
}

func (r *ReconcilerHandler) checkCache(ctx context.Context, namespacedName string, typed client.Object) (error, bool) {
	err := r.Client.Get(ctx, client.ObjectKey{Name: namespacedName}, typed)
	if err != nil {
		if errors.IsNotFound(err) {
			controllerLog.Info("Cache miss", "namespacedName", namespacedName)
			return err, true
		}
		controllerLog.Error(err, "Failed to get object from cache", "namespacedName", namespacedName)
		return err, false
	}

	return nil, false
}
