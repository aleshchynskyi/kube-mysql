/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	kubesqlv1alpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MysqlClusterReconciler reconciles a MysqlCluster object
type MysqlClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlclusters/finalizers,verbs=update

func (r *MysqlClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, r.reconcileWithoutResult(ctx, req)
}
func (r *MysqlClusterReconciler) reconcileWithoutResult(ctx context.Context, req ctrl.Request) error {
	logger := ctrl.LoggerFrom(ctx)
	objectKey := req.NamespacedName
	currentInstance := &kubesqlv1alpha1.MysqlCluster{}
	if err := r.Get(ctx, objectKey, currentInstance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Cluster is getting deleted")
			return nil
		}
		logger.Error(err, "Cannot get cluster CR")
		return err
	}
	cluster := currentInstance.DeepCopy()

	currentConfigSpec, err := r.GetClusterConfig(ctx, cluster)
	if err != nil {
		return err
	}
	cluster.Status.ConfigSpec = currentConfigSpec

	err = r.deployMysqlCluster(ctx, cluster)
	if err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(currentInstance.Status, cluster.Status) {
		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Cannot update cluster status")
			return err
		}
		logger.Info("Updated cluster status")
		currentInstance.Status = cluster.Status
	}

	if !equality.Semantic.DeepEqual(currentInstance, cluster) {
		if err := r.Update(ctx, cluster); err != nil {
			logger.Error(err, "Cannot update cluster CR")
			return err
		}
		logger.Info("Updated cluster CR")
	}

	return nil
}

func (r *MysqlClusterReconciler) deployMysqlCluster(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) error {
	_, err := r.CreateOrUpdatePVC(ctx, cluster)
	if err != nil {
		return err
	}

	_, err = r.createOrUpdateService(ctx, cluster)
	if err != nil {
		return err
	}

	_, err = r.createOrUpdateSet(ctx, cluster)
	if err != nil {
		return err
	}
	return nil
}

func buildLabels(cluster *kubesqlv1alpha1.MysqlCluster) map[string]string {
	return map[string]string{
		"vellanci.gh/mysql-cluster": cluster.Name,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubesqlv1alpha1.MysqlCluster{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
