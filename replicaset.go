package main

import (
	"fmt"
	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewReplicaSet(pod *core_v1.Pod, replicas *int32) *apps_v1.ReplicaSet {
	return &apps_v1.ReplicaSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprintf("%s-rs", pod.Name),
			Namespace: pod.Namespace,
		},
		Spec: apps_v1.ReplicaSetSpec{
			Replicas: replicas,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"app": pod.Labels["app"],
				},
			},
			Template: core_v1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{
						"app": pod.Labels["app"],
					},
				},
				Spec: core_v1.PodSpec{
					Containers: []core_v1.Container{
						{
							Name:  pod.Spec.Containers[0].Name,
							Image: pod.Spec.Containers[0].Image,
						},
					},
				},
			},
		},
	}
}
