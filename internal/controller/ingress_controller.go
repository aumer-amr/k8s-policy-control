package controller

import (
	"context"
	"fmt"

	"github.com/aumer-amr/k8s-policy-control/internal/policy"
	"github.com/go-logr/logr"
	networkingv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ReconcilerHandler) reconcileIngress(ctx context.Context, req ctrl.Request, controllerLog logr.Logger) (ctrl.Result, error) {
	ingress := &networkingv1.Ingress{}
	err, cacheMiss := r.checkCache(ctx, req.NamespacedName.String(), ingress)

	if err != nil {
		if cacheMiss {
			fmt.Println("reconcileIngress cache miss")
			return ctrl.Result{}, nil
		}
		fmt.Println("reconcileIngress error", err)
		return ctrl.Result{}, err
	}

	policy.ApplyPoliciesByType(policy.PolicyTypeIngress, ingress, r.Manager)

	return ctrl.Result{}, nil
}
