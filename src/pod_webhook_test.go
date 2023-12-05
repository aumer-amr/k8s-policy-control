package main

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoveContainerLimits(t *testing.T) {
	podFixture := NewPodFixture("test-pod", "default")

	removeContainerLimits(&podFixture.Spec.Containers[0], v1.ResourceCPU, podFixture)

	cpuLimit, exist := podFixture.Spec.Containers[0].Resources.Limits[v1.ResourceCPU]
	if exist {
		t.Errorf("Expected CPU limit to not exist, value: %v", cpuLimit)
	}
}

func NewPodFixture(name string, namespace string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "myapp",
			},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container-1",
					Image: "nginx:latest",
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("1"),
							v1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Requests: v1.ResourceList{
							v1.ResourceCPU:    resource.MustParse("500m"),
							v1.ResourceMemory: resource.MustParse("256Mi"),
						},
					},
				},
			},
		},
	}
}
