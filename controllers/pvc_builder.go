package controllers

import (
	"context"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildPVC(cluster *v1alpha1.MysqlCluster) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name + "-mysql",
			Namespace: cluster.Namespace,
			Labels:    buildLabels(cluster),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func updatePVC(cluster *v1alpha1.MysqlCluster, pvc *corev1.PersistentVolumeClaim) error {
	if cluster.Spec.Storage.Resources != nil {
		pvc.Spec.Resources = *cluster.Spec.Storage.Resources
	}
	return nil
}

func (r *MysqlClusterReconciler) CreateOrUpdatePVC(ctx context.Context, cluster *v1alpha1.MysqlCluster) (*corev1.PersistentVolumeClaim, error) {
	if cluster.Spec.Storage.VolumeSource != nil {
		return nil, nil
	}

	pvc := buildPVC(cluster)
	if err := r.SetControllerReference(ctx, cluster, pvc); err != nil {
		return nil, err
	}
	_, err := r.CreateOrUpdate(ctx, pvc, func() error {
		return updatePVC(cluster, pvc)
	})
	cluster.Status.StorageVolumeSource = corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: pvc.Name,
		},
	}
	return pvc, err
}
