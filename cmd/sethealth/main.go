package main

import (
	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	metav1ac "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// helper to set an arbitrary status condition on a pod

func main() {
	if err := run(context.Background()); err != nil {
		panic(err)
	}
}

func run(ctx context.Context) error {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("HOME")+"/.kube/config")
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	pods, err := clientset.CoreV1().Pods("default").List(ctx, metav1.ListOptions{
		LabelSelector: "app=server",
	})
	if err != nil {
		return err
	}

	now := time.Now()

	for _, pod := range pods.Items {
		_, err = clientset.CoreV1().Pods(pod.Namespace).ApplyStatus(ctx, &corev1ac.PodApplyConfiguration{
			TypeMetaApplyConfiguration: metav1ac.TypeMetaApplyConfiguration{
				APIVersion: toPtr("v1"),
				Kind:       toPtr("Pod"),
			},
			ObjectMetaApplyConfiguration: &metav1ac.ObjectMetaApplyConfiguration{
				Name: toPtr(pod.Name),
			},
			Status: &corev1ac.PodStatusApplyConfiguration{
				Conditions: []corev1ac.PodConditionApplyConfiguration{
					{
						Type:               toPtr(corev1.PodConditionType("Health")),
						Status:             toPtr(corev1.ConditionTrue),
						LastProbeTime:      &metav1.Time{Time: now},
						LastTransitionTime: &metav1.Time{Time: now},
						Reason:             toPtr("LooksGood"),
						Message:            toPtr("everything looks good"),
					},
				},
			},
		}, metav1.ApplyOptions{
			FieldManager: "sethealth",
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func toPtr[T any](v T) *T {
	return &v
}
