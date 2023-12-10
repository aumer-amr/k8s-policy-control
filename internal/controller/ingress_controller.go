package controller

import (
	"context"
	"fmt"

	"github.com/aumer-amr/k8s-policy-control/internal/policy"
	"github.com/go-logr/logr"
	networkv1 "k8s.io/api/networking/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *ReconcilerHandler) reconcileIngress(ctx context.Context, req ctrl.Request, controllerLog logr.Logger) (ctrl.Result, error) {
	err, cacheMiss := r.checkCache(ctx, req.NamespacedName.String(), &networkv1.Ingress{})

	if err != nil && !cacheMiss {
		return ctrl.Result{}, err
	} else if err != nil && cacheMiss {
		return ctrl.Result{}, nil
	}

	ingress := &networkv1.Ingress{}
	err = r.Client.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		return ctrl.Result{}, err
	}

	fmt.Println(ingress.Labels)
	if ingress.DeletionTimestamp.IsZero() {
		policy.ApplyPoliciesByType(policy.PolicyTypeIngress, ingress, policy.PolicyOperationUpsert)
	} else {
		policy.ApplyPoliciesByType(policy.PolicyTypeIngress, ingress, policy.PolicyOperationDelete)
	}

	return ctrl.Result{}, nil
}
