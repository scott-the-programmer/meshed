package kubernetes

import (
	"errors"
	"meshed/kubernetes/resources"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Environment int

const (
	Staging Environment = iota
	Live
)

func NewCerts(ctx *pulumi.Context, provider *kubernetes.Provider, replacer *resources.Replacer, parent *helm.Release, staging bool) error {

	env := Live
	if staging {
		env = Staging
	}

	var issuer string
	var certificate string
	switch env {
	case Staging:
		issuer = "kubernetes/resources/cert-manager/acme-issuer-staging.template.yaml"
		certificate = "kubernetes/resources/cert-manager/certificate-staging.template.yaml"
	case Live:
		issuer = "kubernetes/resources/cert-manager/acme-issuer.template.yaml"
		certificate = "kubernetes/resources/cert-manager/certificate.template.yaml"
	default:
		return errors.New("unknown environment")
	}

	riss, err := replacer.Replace(issuer)
	if err != nil {
		return err
	}
	rc, err := replacer.Replace(certificate)
	if err != nil {
		return err
	}
	_, err = yaml.NewConfigFile(ctx, "issuer",
		&yaml.ConfigFileArgs{
			File: riss,
		},
		pulumi.Provider(provider), pulumi.Parent(parent),
	)
	if err != nil {
		return err
	}

	_, err = yaml.NewConfigFile(ctx, "cert",
		&yaml.ConfigFileArgs{
			File: rc,
		},
		pulumi.Provider(provider), pulumi.Parent(parent))
	if err != nil {
		return err
	}

	return nil
}
