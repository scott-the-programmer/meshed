package apps

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudflaredArgs struct {
	Image      pulumi.StringPtrInput
	TunnelName pulumi.StringInput
	Subdomain  pulumi.StringInput
	Domain     pulumi.StringInput
}

type CloudflaredSidecarConfig struct {
	Ctx         *pulumi.Context
	Provider    *kubernetes.Provider
	Namespace   *corev1.Namespace
	Name        string
	Args        *CloudflaredArgs
	ServiceName string
	ServicePort int
	SecretName  string
	Hostname    pulumi.StringInput
}

// CloudflaredSidecar provisions the cloudflared ConfigMap and Deployment for a service.
func CloudflaredSidecar(config *CloudflaredSidecarConfig) error {
	image := pulumi.StringPtr("cloudflare/cloudflared:latest-arm64")
	if config.Args.Image != nil {
		image = pulumi.Sprintf("%v", config.Args.Image).ToStringPtrOutput()
	}

	// Create ConfigMap for cloudflared configuration
	cfConfigMap, err := corev1.NewConfigMap(config.Ctx, fmt.Sprintf("%s-cf-config", config.Name), &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(fmt.Sprintf("%s-cf-config", config.Name)),
			Namespace: config.Namespace.Metadata.Name(),
		},
		Data: pulumi.StringMap{
			"config.yaml": pulumi.Sprintf(`
tunnel: %s
credentials-file: /etc/cloudflared/creds/creds.json
metrics: 0.0.0.0:2000
no-autoupdate: true
loglevel: debug
originRequest:
  connectTimeout: 30s
  noTLSVerify: true
ingress:
- hostname: %s
  service: http://%s:%d
- service: http_status:404
`, config.Args.TunnelName, config.Hostname, config.ServiceName, config.ServicePort),
		},
	}, pulumi.Provider(config.Provider))
	if err != nil {
		return err
	}

	_, err = appsv1.NewDeployment(config.Ctx, fmt.Sprintf("%s-cf-deployment", config.Name), &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: config.Namespace.Metadata.Name(),
			Annotations: pulumi.StringMap{
				"keel.sh/policy":       pulumi.String("all"),
				"keel.sh/trigger":      pulumi.String("poll"),
				"keel.sh/pollSchedule": pulumi.String("@every 1m"),
			},
		},
		Spec: appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String(fmt.Sprintf("%s-cf", config.Name)),
				},
			},
			Replicas: pulumi.Int(1),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app":     pulumi.String(fmt.Sprintf("%s-cf", config.Name)),
						"service": pulumi.String(fmt.Sprintf("%s-cf", config.Name)),
					},
					Namespace: config.Namespace.Metadata.Name(),
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Name:  pulumi.String(fmt.Sprintf("%s-cf", config.Name)),
							Image: image,
							Args: pulumi.StringArray{
								pulumi.String("tunnel"),
								pulumi.String("--config"),
								pulumi.String("/etc/cloudflared/config/config.yaml"),
								pulumi.String("run"),
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("config"),
									MountPath: pulumi.String("/etc/cloudflared/config"),
									ReadOnly:  pulumi.Bool(true),
								},
								&corev1.VolumeMountArgs{
									Name:      pulumi.String("creds"),
									MountPath: pulumi.String("/etc/cloudflared/creds"),
									ReadOnly:  pulumi.Bool(true),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Path: pulumi.String("/ready"),
									Port: pulumi.Int(2000),
								},
								InitialDelaySeconds: pulumi.Int(10),
								PeriodSeconds:       pulumi.Int(10),
								FailureThreshold:    pulumi.Int(1),
							},
						}},
					Volumes: &corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("config"),
							ConfigMap: &corev1.ConfigMapVolumeSourceArgs{
								Name: cfConfigMap.Metadata.Name(),
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("creds"),
							Secret: &corev1.SecretVolumeSourceArgs{
								SecretName: pulumi.String(config.SecretName),
							},
						},
					},
				},
			},
		},
	}, pulumi.Provider(config.Provider))
	return err
}
