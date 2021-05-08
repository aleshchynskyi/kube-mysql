package controllers

import (
	"context"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime"
)

func buildService(cluster *v1alpha1.MysqlCluster) *v1.Service {
	return &v1.Service{
		ObjectMeta: controllerruntime.ObjectMeta{
			Name:      cluster.Name + "-mysql",
			Namespace: cluster.Namespace,
			Labels:    buildLabels(cluster),
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Name: "mysql",
					Port: 3306,
				},
			},
			Selector:  buildLabels(cluster),
			ClusterIP: v1.ClusterIPNone,
			Type:      v1.ServiceTypeClusterIP,
		},
	}
}

func updateService(_ *v1alpha1.MysqlCluster, service *v1.Service) error {
	service.Spec.Ports = []v1.ServicePort{
		{
			Name: "mysql",
			Port: 3306,
		},
	}
	return nil
}

func (r *MysqlClusterReconciler) createOrUpdateService(ctx context.Context, cluster *v1alpha1.MysqlCluster) (*v1.Service, error) {
	service := buildService(cluster)
	if err := r.SetControllerReference(ctx, cluster, service); err != nil {
		return nil, err
	}
	_, err := r.CreateOrUpdate(ctx, service, func() error {
		return updateService(cluster, service)
	})
	if err != nil {
		return nil, err
	}
	cluster.Status.Service = service.Name
	return service, err
}
