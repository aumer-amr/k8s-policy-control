package policy

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	policyLog = ctrl.Log.WithName("policy")
)

const (
	PolicyTypeUnknown = iota
	PolicyTypePod
	PolicyTypeIngress
)

var policyRegistry []PolicyInterface

type PolicyInterface interface {
	Name() string
	Validate(obj runtime.Object, mgr ctrl.Manager) (error, bool)
	Apply(obj runtime.Object, mgr ctrl.Manager) error
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

func ApplyPoliciesByType(policyType int, obj runtime.Object, mgr ctrl.Manager) error {
	policies := PoliciesByType(policyType)
	for _, p := range policies {
		policyLog.Info("applying policy", "policy", p.Name())
		err, result := p.Validate(obj, mgr)

		if err != nil {
			policyLog.Error(err, "error validating policy", "policy", p.Name())
			return err
		} else if !result {
			policyLog.Info("policy not applicable", "policy", p.Name())
			return nil
		}

		err = p.Apply(obj, mgr)
		if err != nil {
			policyLog.Error(err, "error running policy", "policy", p.Name())
			return err
		}
	}
	return nil
}

func RegisterPolicies() {
	policies := AllPolicies()
	for _, p := range policies {
		policyLog.Info("registering policy", "policy", p.Name())
	}
}
