package controllers

import (
	"context"
	mysqlalpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

const DefaultMysqlConfig = "default-mysql-config"

func (r *MysqlClusterReconciler) GetClusterConfig(ctx context.Context, cluster *mysqlalpha1.MysqlCluster) (*mysqlalpha1.MysqlConfigSpec, error) {
	defaultMysqlConfig, err := r.GetMysqlConfig(ctx, DefaultMysqlConfig)
	if err != nil {
		return nil, err
	}
	referredMysqlConfig, err := r.GetMysqlConfig(ctx, cluster.Spec.Config.Name)
	if err != nil {
		return nil, err
	}
	currentConfigSpec := MergeConfigs(defaultMysqlConfig, referredMysqlConfig, cluster.Spec.Config.Spec)
	return currentConfigSpec, nil
}

func (r *MysqlClusterReconciler) GetMysqlConfig(ctx context.Context, configName string) (*mysqlalpha1.MysqlConfigSpec, error) {
	config := &mysqlalpha1.MysqlConfig{}
	if configName == "" {
		return &config.Spec, nil
	}
	logger := ctrl.LoggerFrom(ctx)
	if err := r.Get(ctx, types.NamespacedName{Name: configName}, config); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Cannot find config", "name", configName)
			return &config.Spec, err
		}
		return &config.Spec, err
	}
	return &config.Spec, nil
}

func MergeConfigs(configs ...*mysqlalpha1.MysqlConfigSpec) *mysqlalpha1.MysqlConfigSpec {
	var result = &mysqlalpha1.MysqlConfigSpec{}
	for _, next := range configs {
		if next == nil {
			continue
		}
		for name, image := range next.Images {
			result.Images[name] = image
		}
	}
	return result
}
