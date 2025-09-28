package apps

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type BlogApiArgs struct {
	Cloudflared *CloudflaredArgs
}

func NewBlogApi(ctx *pulumi.Context, provider *kubernetes.Provider, ns *corev1.Namespace, name string, args *BlogApiArgs) error {

	appLabels := pulumi.StringMap{
		"app": pulumi.String(name),
	}

	_, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-deployment", name), &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
			Annotations: pulumi.StringMap{
				"keel.sh/policy":       pulumi.String("all"),
				"keel.sh/trigger":      pulumi.String("poll"),
				"keel.sh/pollSchedule": pulumi.String("@every 1m"),
			},
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
							Image: pulumi.String("ghcr.io/scott-the-programmer/blog-api:latest"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(8080),
									Name:          pulumi.String("http"),
									Protocol:      pulumi.String("TCP"),
								},
							},
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
					TargetPort: pulumi.Int(8080),
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
		hostname := pulumi.Sprintf("%s.%s", args.Cloudflared.Subdomain, args.Cloudflared.Domain)
		err = CloudflaredSidecar(&CloudflaredSidecarConfig{
			Ctx:         ctx,
			Provider:    provider,
			Namespace:   ns,
			Name:        name,
			Args:        args.Cloudflared,
			ServiceName: name,
			ServicePort: 80,
			SecretName:  "blog-api-cloudflared-file",
			Hostname:    hostname,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
