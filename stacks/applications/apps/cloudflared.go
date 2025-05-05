package apps

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudflaredArgs struct {
	Image      pulumi.StringPtrInput
	TunnelName pulumi.StringInput
	Subdomain  pulumi.StringInput
	Domain     pulumi.StringInput
}