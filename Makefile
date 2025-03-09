.PHONY: build
build:
	go mod tidy

.PHONY: cluster-preview
cluster-preview:
	cd stacks/cluster && pulumi stack select "cluster" -c && pulumi preview

.PHONY: cluster-up
cluster-up:
	cd stacks/cluster && pulumi stack select "cluster" -c && pulumi up

.PHONY: cluster-down
cluster-down:
	cd stacks/cluster && pulumi stack select "cluster" -c && pulumi destroy -y

.PHONY: mesh-preview
mesh-preview:
	cd stacks/mesh && pulumi stack select "mesh" -c && pulumi preview

.PHONY: mesh-up
mesh-up:
	cd stacks/mesh && pulumi stack select "mesh-local" -c && pulumi up

.PHONY: mesh-down
mesh-down:
	cd stacks/mesh && pulumi stack select "mesh" -c && pulumi destroy -y

.PHONY: apps-preview
apps-preview:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi preview

.PHONY: apps-up
apps-up:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi up

.PHONY: apps-down
apps-down:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi destroy -y

.PHONY: monitoring-preview
monitoring-preview:
	cd stacks/monitoring && pulumi stack select "monitoring" -c && pulumi preview

.PHONY: monitoring-up
monitoring-up:
	cd stacks/monitoring && pulumi stack select "monitoring" -c && pulumi up

.PHONY: monitoring-down
monitoring-down:
	cd stacks/monitoring && pulumi stack select "monitoring" -c && pulumi destroy -y

.PHONY: load-config
load-config:
	cd stacks/cluster && pulumi stack output kubeconfig --show-secrets > $$HOME/.kube/config

.PHONY: grafana
grafana:
	kubectl port-forward -n monitoring svc/grafana 3000:3000

.PHONY: nuke
nuke:
	./kubernetes/resources/nuke.sh
