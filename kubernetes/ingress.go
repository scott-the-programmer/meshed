package kubernetes

import (
	"meshed/kubernetes/resources"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	v1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Container struct {
	Resources map[string]interface{}
}

type PodSpec struct {
	Containers []Container
}

type Template struct {
	Spec PodSpec
}

type DeploymentSpec struct {
	Template Template
}

type Deployment struct {
	Kind     string
	Metadata map[string]interface{}
	Spec     DeploymentSpec
}

func getDeployment(state map[string]interface{}) (*Deployment, bool) {
	kind, ok := state["kind"].(string)
	if !ok {
		return nil, false
	}

	metadata, ok := state["metadata"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	spec, ok := state["spec"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	podSpec, ok := template["spec"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	containers, ok := podSpec["containers"].([]interface{})
	if !ok || len(containers) == 0 {
		return nil, false
	}

	container, ok := containers[0].(map[string]interface{})
	if !ok {
		return nil, false
	}

	resources, ok := container["resources"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	return &Deployment{
		Kind:     kind,
		Metadata: metadata,
		Spec: DeploymentSpec{
			Template: Template{
				Spec: PodSpec{
					Containers: []Container{
						{
							Resources: resources,
						},
					},
				},
			},
		},
	}, true
}

func NewMesh(ctx *pulumi.Context, provider *kubernetes.Provider, replacer *resources.Replacer) (*helm.Release, pulumi.StringPtrOutput, error) {

	istioRepo := "https://istio-release.storage.googleapis.com/charts"
	ri, err := replacer.Replace("kubernetes/resources/ingress.template.yaml")
	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	istioNs, err := v1.NewNamespace(ctx, "istio-system", &v1.NamespaceArgs{
		Metadata: metav1.ObjectMetaArgs{Name: pulumi.String("istio-system")},
	}, pulumi.Provider(provider))

	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	base, err := helm.NewChart(ctx, "istio-base", helm.ChartArgs{
		Chart:     pulumi.String("base"),
		Namespace: istioNs.Metadata.Name().Elem(),
		FetchArgs: helm.FetchArgs{
			Repo: pulumi.String(istioRepo),
		},
	}, pulumi.Provider(provider), pulumi.Parent(istioNs))

	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	discovery, err := helm.NewChart(ctx, "istiod", helm.ChartArgs{
		Chart:     pulumi.String("istiod"),
		Namespace: istioNs.Metadata.Name().Elem(),
		FetchArgs: helm.FetchArgs{
			Repo: pulumi.String(istioRepo),
		},
		Transformations: []yaml.Transformation{
			// Lower the default request value for the istiod container
			func(state map[string]interface{}, opts ...pulumi.ResourceOption) {
				deployment, ok := getDeployment(state)
				if ok && deployment.Metadata["name"] == "istiod" {
					deployment.Spec.Template.Spec.Containers[0].Resources["requests"] = map[string]interface{}{"memory": "1024Mi"}
				}
			},
		}}, pulumi.Provider(provider), pulumi.Parent(base))

	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	istioGwNs, err := v1.NewNamespace(ctx, "istio-ingress", &v1.NamespaceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Labels: pulumi.StringMap{
				"istio-injection": pulumi.String("enabled"),
			},
			Name: pulumi.String("istio-ingress"),
		},
	}, pulumi.Provider(provider), pulumi.Parent(discovery))

	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}
	gw, err := helm.NewRelease(ctx, "istio-gateway", &helm.ReleaseArgs{
		Chart:     pulumi.String("gateway"),
		Namespace: istioGwNs.Metadata.Name().Elem(),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String(istioRepo),
		},
		Name:        pulumi.String("istio-ingress"),
		Timeout:     pulumi.Int(600),
		WaitForJobs: pulumi.Bool(true),
		Atomic:      pulumi.Bool(true),
		Values: pulumi.Map{
				"_internal_defaults_do_not_set": pulumi.Map{ // BRING IT ON
					"limits": pulumi.Map{
						"cpu": pulumi.String("500m"),
					},
				},
			},
	}, pulumi.Provider(provider), pulumi.Parent(istioGwNs))
	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	service, err := v1.GetService(ctx, "istio-ingress/istio-gateway", gw.ID(), nil, pulumi.Provider(provider), pulumi.Parent(gw))
	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	loadBalancerIp := service.Status.ApplyT(func(status *v1.ServiceStatus) *string {
		if status == nil || status.LoadBalancer == nil || len(status.LoadBalancer.Ingress) == 0 {
			return nil
		}
		return status.LoadBalancer.Ingress[0].Ip
	}).(pulumi.StringPtrOutput)

	loadBalancerStringPtr := loadBalancerIp.ApplyT(func(ip interface{}) *string {
		return ip.(*string)
	}).(pulumi.StringPtrOutput)

	_, err = yaml.NewConfigFile(ctx, "ingress",
		&yaml.ConfigFileArgs{
			File: ri,
		},
		pulumi.Provider(provider), pulumi.Parent(service),
	)
	if err != nil {
		return nil, pulumi.StringPtrOutput{}, err
	}

	return gw, loadBalancerStringPtr, nil
}
