package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type CommonReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CommonReconciler) CreateOrUpdate(ctx context.Context, object client.Object, fn func() error) (controllerutil.OperationResult, error) {
	logger := r.loggerForObject(ctx, object)
	result, err := ctrl.CreateOrUpdate(ctx, r.Client, object, fn)
	if err != nil {
		logger.Error(err, "Operation failed", "result", result)
		return result, err
	}
	if result != controllerutil.OperationResultNone {
		logger.Info("Operation succeeded", "result", result)
	}
	return result, nil
}

func (r *CommonReconciler) loggerForObject(ctx context.Context, object client.Object) logr.Logger {
	kinds, _, err := r.Scheme.ObjectKinds(object)
	if len(kinds) < 1 {
		logger := ctrl.LoggerFrom(ctx, "name", object.GetName())
		logger.Error(err, "Cannot find a kind of an object")
		return logger
	}
	return ctrl.LoggerFrom(ctx,
		"kind", kinds[0].Kind,
		"name", object.GetName(),
	)
}

func (r CommonReconciler) SetControllerReference(ctx context.Context, owner client.Object, object client.Object) error {
	if err := ctrl.SetControllerReference(owner, object, r.Scheme); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "Cannot set controller reference")
		return err
	}
	return nil
}
