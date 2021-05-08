package controllers

import (
	"context"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *MysqlClusterReconciler) CreateOrUpdate(ctx context.Context, object client.Object, fn func() error) (controllerutil.OperationResult, error) {
	logger := controllerruntime.LoggerFrom(ctx,
		"kind", object.GetObjectKind().GroupVersionKind().Kind,
		"name", object.GetName(),
	)
	result, err := controllerruntime.CreateOrUpdate(ctx, r.Client, object, fn)
	if err != nil {
		logger.Error(err, "Operation failed", "result", result)
		return result, err
	}
	logger.Info("Operation succeeded", "result", result)
	return result, nil
}
