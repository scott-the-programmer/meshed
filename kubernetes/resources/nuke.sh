#!/bin/bash
set -e

# Deleting all resources in all namespaces except kube-system and cert-manager
for ns in $(kubectl get ns --no-headers | awk '{print $1}' | grep -vE '(^kube-system$|^cert-manager$|^default$|^kube-public$|^kube-node-lease$)'); do
  kubectl proxy &
  proxy_pid=$!
  sleep 1
  curl -H "Content-Type: application/json" -X PUT --data '{"apiVersion":"v1","kind":"Namespace","metadata":{"name":"'$ns'"},"spec":{"finalizers":[]}}' http://localhost:8001/api/v1/namespaces/$ns/finalize
  kill $proxy_pid
  kubectl delete ns $ns --grace-period=0 --force
done

# Deleting all CRDs
for crd in $(kubectl get crd --no-headers | awk '{print $1}'); do
  kubectl delete crd $crd
done
