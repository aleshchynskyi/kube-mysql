package controllers

import (
	"context"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func buildStatefulSet(cluster *v1alpha1.MysqlCluster) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      cluster.Name + "-mysql",
			Namespace: cluster.Namespace,
			Labels:    buildLabels(cluster),
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: ctrl.ObjectMeta{
					Labels: buildLabels(cluster),
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: buildLabels(cluster),
			},
		},
	}
}

func updateStatefulSet(cluster *v1alpha1.MysqlCluster, set *appsv1.StatefulSet) error {
	set.Spec.Replicas = &cluster.Spec.Replicas
	set.Spec.Template.Spec = buildPodSpec(cluster)
	set.Spec.ServiceName = cluster.Status.Service
	return nil
}

func buildPodSpec(cluster *v1alpha1.MysqlCluster) corev1.PodSpec {
	return corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "mysql",
				Image: cluster.Status.ConfigSpec.Images["mysql"],
				VolumeMounts: []corev1.VolumeMount{
					{Name: "storage", MountPath: "/var/lib/mysql"},
				},
				Ports: []corev1.ContainerPort{
					{Name: "mysql", ContainerPort: 3306},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "MYSQL_ROOT_PASSWORD",
						Value: "root",
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "storage",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: cluster.Status.StoragePVC,
					},
				},
			},
		},
	}
}

func (r *MysqlClusterReconciler) createOrUpdateSet(ctx context.Context, cluster *v1alpha1.MysqlCluster) (*appsv1.StatefulSet, error) {
	statefulSet := buildStatefulSet(cluster)
	if err := r.SetControllerReference(ctx, cluster, statefulSet); err != nil {
		return nil, err
	}
	_, err := r.CreateOrUpdate(ctx, statefulSet, func() error {
		return updateStatefulSet(cluster, statefulSet)
	})
	if err != nil {
		return nil, err
	}
	cluster.Status.StatefulSet = statefulSet.Name
	return statefulSet, nil
}
