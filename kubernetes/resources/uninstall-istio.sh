#!/bin/bash

helm delete istio-system -n istio-system
helm delete istiod -n istio-system
helm delete istio-base -n istio-system

kubectl get crd -oname | grep --color=never 'istio.io' | xargs kubectl delete