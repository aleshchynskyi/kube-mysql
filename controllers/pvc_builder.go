package controllers

import (
	"context"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildPVC(cluster *v1alpha1.MysqlCluster) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: v12.ObjectMeta{
			Name:      cluster.Name + "-mysql",
			Namespace: cluster.Namespace,
			Labels:    buildLabels(cluster),
		},
	}
}

func updatePVC(_ *v1alpha1.MysqlCluster, pvc *v1.PersistentVolumeClaim) error {
	pvc.Spec.Resources = v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{
			v1.ResourceStorage: resource.MustParse("1Gi"),
		},
	}
	return nil
}

func (r *MysqlClusterReconciler) CreateOrUpdatePVC(ctx context.Context, cluster *v1alpha1.MysqlCluster) (*v1.PersistentVolumeClaim, error) {
	pvc := buildPVC(cluster)
	_, err := r.CreateOrUpdate(ctx, pvc, func() error {
		return updatePVC(cluster, pvc)
	})
	cluster.Status.StoragePVC = pvc.Name
	return pvc, err
}
