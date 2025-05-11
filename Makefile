.PHONY: build
build:
	go mod tidy

.PHONY: apps-preview
apps-preview:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi preview

.PHONY: apps-up
apps-up:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi up

.PHONY: apps-down
apps-down:
	cd stacks/applications && pulumi stack select "applications" -c && pulumi destroy -y

.PHONY: load-config
load-config:
	cd stacks/cluster && pulumi stack output kubeconfig --show-secrets > $$HOME/.kube/config

.PHONY: grafana
grafana:
	kubectl port-forward -n monitoring svc/grafana 3000:3000

.PHONY: nuke
nuke:
	./kubernetes/resources/nuke.sh

.PHONY: ingress-ip
ingress-ip:
	kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
