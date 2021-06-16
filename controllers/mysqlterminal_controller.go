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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kubesqlv1alpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
)

// MysqlTerminalReconciler reconciles a MysqlTerminal object
type MysqlTerminalReconciler struct {
	CommonReconciler
}

//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlterminals,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlterminals/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubesql.vellanci.gh,resources=mysqlterminals/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MysqlTerminal object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MysqlTerminalReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	return ctrl.Result{}, r.reconcileWithoutResult(ctx, req)
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlTerminalReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kubesqlv1alpha1.MysqlTerminal{}).
		Owns(&corev1.Secret{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}

func (r *MysqlTerminalReconciler) reconcileWithoutResult(ctx context.Context, req ctrl.Request) error {
	logger := ctrl.LoggerFrom(ctx)
	objectKey := req.NamespacedName
	currentInstance := &kubesqlv1alpha1.MysqlTerminal{}
	if err := r.Get(ctx, objectKey, currentInstance); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Terminal is getting deleted")
			return nil
		}
		logger.Error(err, "Failed to get terminal CR")
		return err
	}
	terminal := currentInstance.DeepCopy()

	cluster := &kubesqlv1alpha1.MysqlCluster{}
	clusterId := client.ObjectKey{
		Namespace: terminal.Spec.Cluster.Namespace,
		Name:      terminal.Spec.Cluster.Name,
	}
	if clusterId.Namespace == "" {
		clusterId.Namespace = terminal.Namespace
	}
	if err := r.Get(ctx, clusterId, cluster); err != nil {
		if errors.IsNotFound(err) {
			logger.Error(err, "Cluster not found", "id", clusterId)
			return err
		}
		logger.Error(err, "Failed to get MySQL cluster")
		return err
	}

	secret, err := r.deployTerminalConf(ctx, cluster, terminal)
	if err != nil {
		return err
	}

	if err := r.deployTerminalPod(ctx, cluster, secret, terminal); err != nil {
		return err
	}

	if !equality.Semantic.DeepEqual(currentInstance.Status, terminal.Status) {
		if err := r.Status().Update(ctx, terminal); err != nil {
			logger.Error(err, "Cannot update terminal status")
			return err
		}
		logger.Info("Updated terminal status")
		currentInstance.Status = terminal.Status
	}

	return nil
}
