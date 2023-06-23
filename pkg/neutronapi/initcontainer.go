package neutronapi

import (
	corev1 "k8s.io/api/core/v1"
)

// InitContainer information
type InitContainer struct {
	ContainerImage       string
	Database             string
	DatabaseHost         string
	DatabaseUser         string
	NeutronSecret        string
	TransportURLSecret   string
	DBPasswordSelector   string
	UserPasswordSelector string
	VolumeMounts         []corev1.VolumeMount
}

// GetInitContainer - init container for neutron services
func GetInitContainer(init InitContainer) []corev1.Container {
	envs := []corev1.EnvVar{
		{
			Name:  "DatabaseHost",
			Value: init.DatabaseHost,
		},
		{
			Name:  "Database",
			Value: init.Database,
		},
		{
			Name:  "DatabaseUser",
			Value: init.DatabaseUser,
		},
		{
			Name: "DatabasePassword",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: init.NeutronSecret,
					},
					Key: init.DBPasswordSelector,
				},
			},
		},
		{
			Name: "NeutronPassword",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: init.NeutronSecret,
					},
					Key: init.UserPasswordSelector,
				},
			},
		},
	}

	if init.TransportURLSecret != "" {
		envTransport := corev1.EnvVar{
			Name: "TransportURL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: init.TransportURLSecret,
					},
					Key: "transport_url",
				},
			},
		}
		envs = append(envs, envTransport)
	}

	return []corev1.Container{
		{
			Name:  "init",
			Image: init.ContainerImage,
			Command: []string{
				"/bin/bash", "-c", "/usr/local/bin/container-scripts/init.sh",
			},
			Env:          envs,
			VolumeMounts: init.VolumeMounts,
		},
	}
}
