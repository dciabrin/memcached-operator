package memcached

import (
	labels "github.com/openstack-k8s-operators/lib-common/modules/common/labels"
	memcachedv1 "github.com/openstack-k8s-operators/memcached-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Deployment returns a Deployment resource for the Memcached CR
func Deployment(m *memcachedv1.Memcached) *appsv1.Deployment {
	matchls := map[string]string{
		"app":   "memcached",
		"cr":    "memcached-" + m.Name,
		"owner": "memcached-operator",
	}
	ls := labels.GetLabels(m, "memcached", matchls)
	replicas := m.Spec.Replicas
	runAsUser := int64(0)

	livenessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       3,
		InitialDelaySeconds: 3,
	}
	readinessProbe := &corev1.Probe{
		// TODO might need tuning
		TimeoutSeconds:      5,
		PeriodSeconds:       5,
		InitialDelaySeconds: 5,
	}

	// TODO might want to disable probes in 'Debug' mode
	livenessProbe.TCPSocket = &corev1.TCPSocketAction{
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(11211)},
	}
	readinessProbe.TCPSocket = &corev1.TCPSocketAction{
		Port: intstr.IntOrString{Type: intstr.Int, IntVal: int32(11211)},
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: matchls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: "memcached-operator-memcached",
					Containers: []corev1.Container{{
						Image:   m.Spec.ContainerImage,
						Name:    "memcached",
						Command: []string{"/usr/bin/dumb-init", "--", "/usr/local/bin/kolla_start"},
						SecurityContext: &corev1.SecurityContext{
							RunAsUser: &runAsUser,
						},
						Env: []corev1.EnvVar{{
							Name:  "KOLLA_CONFIG_STRATEGY",
							Value: "COPY_ALWAYS",
						}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/var/lib/kolla/config_files/src",
							ReadOnly:  true,
							Name:      "config-data",
						}, {
							MountPath: "/var/lib/kolla/config_files",
							ReadOnly:  true,
							Name:      "kolla-config",
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
						ReadinessProbe: readinessProbe,
						LivenessProbe:  livenessProbe,
					}},
					Volumes: []corev1.Volume{
						{
							Name: "kolla-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.Name + "-config-data",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "config.json",
											Path: "config.json",
										},
									},
								},
							},
						},
						{
							Name: "config-data",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: m.Name + "-config-data",
									},
									Items: []corev1.KeyToPath{
										{
											Key:  "memcached",
											Path: "etc/sysconfig/memcached",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return dep
}
