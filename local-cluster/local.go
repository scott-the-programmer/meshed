package linode

import (
	"io/ioutil"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//Returns the current kubeconfig
func NewLocalCluster(ctx *pulumi.Context, name string, version string) (pulumi.StringOutput, error) {
	kubeConfig, err := ioutil.ReadFile("~/.kube/config")
	pulumiStr := pulumi.String(string(kubeConfig))
	return pulumiStr.ToStringOutput(), err
}
