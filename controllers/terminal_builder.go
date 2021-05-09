package controllers

import (
	"context"
	kubesqlv1alpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const MysqlClientConfigName = ".mylogin.cnf"

func (r *MysqlClusterReconciler) deployTerminal(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) error {
	secret, err := r.deployTerminalConf(ctx, cluster)
	if err != nil {
		return err
	}

	if err := r.deployTerminalPod(ctx, cluster, secret); err != nil {
		return err
	}

	return nil
}
func (r *MysqlClusterReconciler) deployTerminalConf(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      cluster.Name + "-terminal-conf",
			Namespace: cluster.Namespace,
		},
	}
	if err := r.SetControllerReference(ctx, cluster, secret); err != nil {
		return nil, err
	}
	_, err := r.CreateOrUpdate(ctx, secret, func() error {
		secret.StringData = map[string]string{
			MysqlClientConfigName: "[client]" +
				"\nuser = " + "root" +
				"\npassword = " + "root" +
				"\nhost = " + cluster.Status.Service,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return secret, nil
}

func (r *MysqlClusterReconciler) deployTerminalPod(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster, secret *corev1.Secret) error {
	pod := buildPod(cluster, secret)
	if err := r.SetControllerReference(ctx, cluster, pod); err != nil {
		return err
	}
	_, err := r.CreateOrUpdate(ctx, pod, func() error {
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *MysqlClusterReconciler) undeployTerminal(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster) error {
	template := buildPod(cluster, nil)
	if err := r.SetControllerReference(ctx, cluster, template); err != nil {
		return err
	}
	logger := ctrl.LoggerFrom(ctx, "terminal-pod", cluster.Name)
	foundPod := &corev1.Pod{}
	key := client.ObjectKey{Namespace: template.Namespace, Name: template.Name}
	if err := r.Get(ctx, key, foundPod); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		logger.Error(err, "Failed to get pod")
		return err
	}
	if !equality.Semantic.DeepEqual(foundPod.OwnerReferences, template.OwnerReferences) {
		// If pod doesn't belong to terminal ignore
		return nil
	}
	if foundPod.ObjectMeta.DeletionTimestamp != nil {
		// If pod is being terminated ignore
		return nil
	}
	logger.Info("Terminal is disabled but pod exists, removing")
	if err := r.Delete(ctx, foundPod); err != nil {
		logger.Error(err, "Failed to delete pod")
		return err
	}
	return nil
}

func (r *MysqlClusterReconciler) buildConfSecret(cluster *kubesqlv1alpha1.MysqlCluster) (*corev1.Secret, error) {
	return &corev1.Secret{StringData: map[string]string{
		MysqlClientConfigName: "[client]" +
			"\nuser = " + "root" +
			"\npassword = " + "root" +
			"\nhost = " + cluster.Status.Service,
	}}, nil
}

func buildPod(cluster *kubesqlv1alpha1.MysqlCluster, secret *corev1.Secret) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      cluster.Name + "-terminal",
			Namespace: cluster.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "mysql",
					Image: cluster.Status.ConfigSpec.Images["mysql"],
					Env: []corev1.EnvVar{
						{
							Name:  "MYSQL_HOST",
							Value: cluster.Status.Service,
						},
					},
					Args: []string{"sleep", "infinity"},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "config",
							MountPath: "/root",
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: secret.Name,
						},
					},
				},
			},
		},
	}
}
