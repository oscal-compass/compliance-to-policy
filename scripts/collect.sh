#!/bin/bash

pushd collect

git pull
git merge origin/main -m "Mearge at `date`"

kubectl get policy -A -o yaml > ./policy-results/00.policies.yaml
kubectl get policySet -A -o yaml > ./policy-results/00.policysets.yaml
kubectl get placementdecisions -A -o yaml > ./policy-results/00.placementdecisions.yaml

pushd policy-results
yq '.items[]' 00.policies.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
yq '.items[]' 00.policysets.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
yq '.items[]' 00.placementdecisions.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
popd

git add policy-results/* 
git commit --allow-empty -m "Update at `date`"
git push

popd