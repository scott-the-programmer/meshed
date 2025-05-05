package apps

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type TermNzArgs struct {
	Cloudflared *CloudflaredArgs
}

func NewTermNz(ctx *pulumi.Context, provider *kubernetes.Provider, ns *corev1.Namespace, name string, args *TermNzArgs) error {

	appLabels := pulumi.StringMap{
		"app": pulumi.String(name),
	}
	_, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-deployment", name), &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
		},
		Spec: appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Replicas: pulumi.Int(1),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels:    appLabels,
					Namespace: ns.Metadata.Name(),
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Name:  pulumi.String(name),
							Image: pulumi.String("ghcr.io/scott-the-programmer/term.nz/term-nz:latest"),
						}},
				},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, fmt.Sprintf("%s-svc", name), &corev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(name),
			Namespace: ns.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":     pulumi.String(name),
				"service": pulumi.String(name),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Ports: &corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),
					TargetPort: pulumi.Int(80),
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: appLabels,
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}
	
	if args != nil && args.Cloudflared != nil {
		cfName := fmt.Sprintf("%s-cf", name)
		image := pulumi.StringPtr("cloudflare/cloudflared:latest")
		if args.Cloudflared.Image != nil {
			image = pulumi.Sprintf("%v", args.Cloudflared.Image).ToStringPtrOutput()
		}

		// Create ConfigMap for cloudflared configuration
		cfConfigMap, err := corev1.NewConfigMap(ctx, fmt.Sprintf("%s-cf-config", name), &corev1.ConfigMapArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Name:      pulumi.String(fmt.Sprintf("%s-cf-config", name)),
				Namespace: ns.Metadata.Name(),
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
  service: http://%s:80
- service: http_status:404
`, args.Cloudflared.TunnelName, args.Cloudflared.Domain, name),
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		cfDeployment, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-cf-deployment", name), &appsv1.DeploymentArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name(),
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"app": pulumi.String(cfName),
					},
				},
				Replicas: pulumi.Int(1),
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: pulumi.StringMap{
							"app":     pulumi.String(cfName),
							"service": pulumi.String(cfName),
						},
						Namespace: ns.Metadata.Name(),
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String(cfName),
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
									SecretName: pulumi.String("termnz-cloudflared-file"),
								},
							},
						},
					},
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		_, err = corev1.NewService(ctx, fmt.Sprintf("%s-cf-svc", name), &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(cfName),
				Namespace: ns.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app":     pulumi.String(cfName),
					"service": pulumi.String(cfName),
				},
			},
			Spec: &corev1.ServiceSpecArgs{
				Type: pulumi.String("ClusterIP"),
				Ports: &corev1.ServicePortArray{
					&corev1.ServicePortArgs{
						Port:       pulumi.Int(8080),
						TargetPort: pulumi.Int(8080),
						Protocol:   pulumi.String("TCP"),
					},
				},
				Selector: pulumi.StringMap{
					"app": pulumi.String(cfName),
				},
			},
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{cfDeployment}))
		if err != nil {
			return err
		}
	}
	
	return nil
}
