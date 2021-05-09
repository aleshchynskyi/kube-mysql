package controllers

import (
	"context"
	kubesqlv1alpha1 "github.com/vellanci/kube-mysql.git/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const MysqlClientConfigName = ".mylogin.cnf"

func (r *MysqlTerminalReconciler) deployTerminalConf(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster, terminal *kubesqlv1alpha1.MysqlTerminal) (*corev1.Secret, error) {
	secret := &corev1.Secret{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      terminal.Name + "-conf",
			Namespace: terminal.Namespace,
		},
	}
	if err := r.SetControllerReference(ctx, terminal, secret); err != nil {
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

func (r *MysqlTerminalReconciler) deployTerminalPod(ctx context.Context, cluster *kubesqlv1alpha1.MysqlCluster, secret *corev1.Secret, terminal *kubesqlv1alpha1.MysqlTerminal) error {
	pod := buildPod(cluster, secret, terminal)
	if err := r.SetControllerReference(ctx, terminal, pod); err != nil {
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

func (r *MysqlTerminalReconciler) buildConfSecret(cluster *kubesqlv1alpha1.MysqlCluster) (*corev1.Secret, error) {
	return &corev1.Secret{StringData: map[string]string{
		MysqlClientConfigName: "[client]" +
			"\nuser = " + "root" +
			"\npassword = " + "root" +
			"\nhost = " + cluster.Status.Service + "." + cluster.Namespace,
	}}, nil
}

func buildPod(cluster *kubesqlv1alpha1.MysqlCluster, secret *corev1.Secret, terminal *kubesqlv1alpha1.MysqlTerminal) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      terminal.Name + "-terminal",
			Namespace: terminal.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "mysql",
					Image: cluster.Status.ConfigSpec.Images["mysql"],
					Env: []corev1.EnvVar{
						{
							Name:  "MYSQL_HOST",
							Value: cluster.Status.Service + "." + cluster.Namespace,
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
