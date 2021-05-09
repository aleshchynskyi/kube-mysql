package controllers

import (
	"context"
	kubesqlv1alpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *MysqlClusterReconciler) deployTerminal(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) error {
	terminal := buildTerminal(cluster)
	if err := r.SetControllerReference(ctx, cluster, terminal); err != nil {
		return err
	}
	_, err := r.CreateOrUpdate(ctx, terminal, func() error {
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func buildTerminal(cluster *kubesqlv1alpha1.MysqlCluster) *kubesqlv1alpha1.MysqlTerminal {
	return &kubesqlv1alpha1.MysqlTerminal{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: kubesqlv1alpha1.MysqlTerminalSpec{
			Cluster: kubesqlv1alpha1.MysqlTerminalCluster{
				Name: cluster.Name,
			},
		},
	}
}

func (r *MysqlClusterReconciler) undeployTerminal(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) error {
	template := buildTerminal(cluster)
	if err := r.SetControllerReference(ctx, cluster, template); err != nil {
		return err
	}
	logger := controllerruntime.LoggerFrom(ctx, "terminal", template.Name)
	found := &kubesqlv1alpha1.MysqlTerminal{}
	key := client.ObjectKey{Namespace: template.Namespace, Name: template.Name}
	if err := r.Get(ctx, key, found); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		logger.Error(err, "Failed to get terminal")
		return err
	}
	if !equality.Semantic.DeepEqual(found.OwnerReferences, template.OwnerReferences) {
		// If terminal doesn't belong to cluster ignore
		return nil
	}
	if found.ObjectMeta.DeletionTimestamp != nil {
		// If terminal is being terminated ignore
		return nil
	}
	logger.Info("Terminal per cluster is disabled but terminal object exists, removing")
	if err := r.Delete(ctx, found); err != nil {
		logger.Error(err, "Failed to delete terminal")
		return err
	}
	return nil
}
