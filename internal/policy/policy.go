package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
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

var policyRegistry []PolicyInterface

type PolicyInterface interface {
	Name() string
	Validate(obj runtime.Object) (error, bool)
	Run(obj runtime.Object) error
	Type() int
}

func RegisterPolicy(impl PolicyInterface) {
	policyRegistry = append(policyRegistry, impl)
}

func GetAllPolicies() []PolicyInterface {
	return policyRegistry
}

func GetPoliciesByType(policyType int) []PolicyInterface {
	var policies []PolicyInterface
	for _, p := range policyRegistry {
		if p.Type() == policyType {
			policies = append(policies, p)
		}
	}
	return policies
}

func ApplyPoliciesByType(policyType int, obj runtime.Object) error {
	policies := GetPoliciesByType(policyType)
	for _, p := range policies {
		policyLog.Info("applying policy", "policy", p.Name())
		err, result := p.Validate(obj)

		if err != nil {
			return err
		} else if !result {
			return nil
		}

		err = p.Run(obj)
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
	}

	return nil, nil
}

func RegisterPolicies() {
	policies := GetAllPolicies()
	for _, p := range policies {
		policyLog.Info("registering policy", "policy", p.Name())
	}
}
