package controllers

import (
	"context"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MysqlClusterReconciler) deployTerminal(ctx context.Context, cluster *v1alpha1.MysqlCluster) error {
	pod := &corev1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name: cluster.Name + "-terminal",
			Namespace: cluster.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "mysql",
					Image: cluster.Status.ConfigSpec.Images["mysql"],
					Env: []corev1.EnvVar{
						{
							Name:  "MYSQL_ROOT_PASSWORD",
							Value: "root",
						},
					},
				},
			},
		},
	}
	if err := r.SetControllerReference(ctx, cluster, pod); err != nil {
		return err
	}
	_, err := r.CreateOrUpdate(ctx, pod, func() error {
		return nil
	})
	if err != nil {
		return err
	}
	return err
}
