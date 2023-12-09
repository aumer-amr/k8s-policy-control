package policy

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	KeepLimitsAnnotation = "policy-control.aumer.io/keep-limits"
	log                  = ctrl.Log.WithName("pod_cpu_limits")
)

type PodCPULimits struct{}

func (p PodCPULimits) Name() string {
	return "Pod CPU Limits"
}

func (p PodCPULimits) Type() int {
	return PolicyTypePod
}

func (p PodCPULimits) Validate(obj runtime.Object) (error, bool) {
	err, result := ValidateByType(PolicyTypePod, obj)
	if err != nil {
		return err, false
	}

	if pod, ok := result.(*corev1.Pod); ok {
		if val, ok := pod.Annotations[KeepLimitsAnnotation]; ok {
			if val == "true" {
				log.Info("Skipping Pod because annotation is explicitly true", "pod", getPodName(pod))
				return nil, false
			}
		}
	} else {
		return fmt.Errorf("could not cast object to Pod"), false
	}

	return nil, true
}

func (p PodCPULimits) Run(obj runtime.Object) error {
	err, result := ValidateByType(PolicyTypePod, obj)
	if err != nil {
		return err
	}

	if pod, ok := result.(*corev1.Pod); ok {
		for _, container := range pod.Spec.InitContainers {
			removeContainerLimits(&container, corev1.ResourceCPU, pod)
		}

		for _, container := range pod.Spec.Containers {
			removeContainerLimits(&container, corev1.ResourceCPU, pod)
		}
	}

	return nil
}

func removeContainerLimits(container *corev1.Container, limitType corev1.ResourceName, pod *corev1.Pod) {
	limits := container.Resources.Limits
	_, cpuLimitExists := limits[limitType]
	if cpuLimitExists {
		delete(limits, limitType)
		log.Info("Removed resource limit",
			"namespace", pod.Namespace,
			"pod", getPodName(pod),
			"container", container.Name,
			"limit", limitType,
		)
	}
}

func getPodName(pod *corev1.Pod) string {
	podName := pod.GetName()
	if podName != "" {
		return podName
	}
	return pod.GetGenerateName()
}

func init() {
	RegisterPolicy(&PodCPULimits{})
}
