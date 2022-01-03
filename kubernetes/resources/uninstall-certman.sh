#!/bin/bash

helm uninstall cert-manager -n cert-manager
kubectl delete roles cert-manager-startupapicheck:create-cert -n cert-manager;
kubectl delete serviceaccount cert-manager-startupapicheck -n cert-manager;
kubectl delete serviceaccount default -n cert-manager;
kubectl delete jobs cert-manager-startupapicheck -n cert-manager;
kubectl delete rolebindings cert-manager-startupapicheck:create-cert -n cert-manager;