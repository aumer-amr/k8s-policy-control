package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	policyLog = ctrl.Log.WithName("policy")
)

const (
	PolicyTypePod = iota
	PolicyTypeIngress
)

const (
	PolicyOperationUpsert = iota
	PolicyOperationDelete
)

var policyRegistry []PolicyInterface

type PolicyInterface interface {
	Name() string
	Validate(obj runtime.Object) (error, bool)
	Apply(obj runtime.Object, operation int) error
	Type() int
}

func RegisterPolicy(impl PolicyInterface) {
	policyRegistry = append(policyRegistry, impl)
}

func AllPolicies() []PolicyInterface {
	return policyRegistry
}

func PoliciesByType(policyType int) []PolicyInterface {
	var policies []PolicyInterface
	for _, p := range policyRegistry {
		if p.Type() == policyType {
			policies = append(policies, p)
		}
	}
	return policies
}

func ApplyPoliciesByType(policyType int, obj runtime.Object, operation int) error {
	policies := PoliciesByType(policyType)
	for _, p := range policies {
		policyLog.Info("applying policy", "policy", p.Name())
		err, result := p.Validate(obj)

		if err != nil {
			return err
		} else if !result {
			return nil
		}

		err = p.Apply(obj, operation)
		if err != nil {
			policyLog.Error(err, "error running policy", p.Name())
			return err
		}
	}
	return nil
}

func ValidateByType(policyType int, obj runtime.Object) (error, interface{}) {
	if policyType == PolicyTypePod {
		pod, ok := obj.(*corev1.Pod)
		if !ok {
			policyLog.Error(nil, "expected a Pod but got a %T", obj)
			return fmt.Errorf("expected a Pod but got a %T", obj), nil
		}
		return nil, pod
	} else if policyType == PolicyTypeIngress {
		ingress, ok := obj.(*networkv1.Ingress)
		if !ok {
			policyLog.Error(nil, "expected an Ingress but got a %T", obj)
			return fmt.Errorf("expected an Ingress but got a %T", obj), nil
		}
		return nil, ingress
	}

	return nil, nil
}

func RegisterPolicies() {
	policies := AllPolicies()
	for _, p := range policies {
		policyLog.Info("registering policy", "policy", p.Name())
	}
}
