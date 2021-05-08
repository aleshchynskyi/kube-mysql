package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/vellanci/kube-mysql.git/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MysqlClusterReconciler) CreateOrUpdate(ctx context.Context, object client.Object, fn func() error) (controllerutil.OperationResult, error) {
	logger := loggerForObject(ctx, object)
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, object, fn)
	if err != nil {
		logger.Error(err, "Operation failed", "result", result)
		return result, err
	}
	logger.Info("Operation succeeded", "result", result)
	return result, nil
}

func loggerForObject(ctx context.Context, object client.Object) logr.Logger {
	return ctrl.LoggerFrom(ctx,
		"kind", object.GetObjectKind().GroupVersionKind().Kind,
		"name", object.GetName(),
	)
}

func (r MysqlClusterReconciler) SetControllerReference(ctx context.Context, cluster *v1alpha1.MysqlCluster, object client.Object) error {
	if err := ctrl.SetControllerReference(cluster, object, r.Scheme); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "Cannot set controller reference")
		return err
	}
	return nil
}
