#!/bin/bash

set -e

kind create cluster --name test-komodor

helm install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace --wait

kubectl apply -f examples/provider/provider.yaml --namespace crossplane-system

kubectl create secret generic komodor-api-key --from-literal=apiKey=${KOMODOR_API_KEY} --namespace crossplane-system

kubectl apply -f examples/provider/providerconfig.yaml --namespace crossplane-system

kubectl apply -f examples/sample/realtimemonitor.yaml --namespace crossplane-system


