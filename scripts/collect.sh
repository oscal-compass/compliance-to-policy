#!/bin/bash

resultdir=./ocm-policy-statuses

pushd collect

git pull

rm -rf  $resultdir
mkdir $resultdir
kubectl get policy -A -o yaml > $resultdir/00.policies.yaml
kubectl get policySet -A -o yaml > $resultdir/00.policysets.yaml
kubectl get placementdecisions -A -o yaml > $resultdir/00.placementdecisions.yaml

pushd $resultdir
yq '.items[]' 00.policies.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
yq '.items[]' 00.policysets.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
yq '.items[]' 00.placementdecisions.yaml -s '.kind + "." + .metadata.namespace + "." + .metadata.name'
popd

git add $resultdir 
git commit --allow-empty -m "Update at `date`"
git push

popd