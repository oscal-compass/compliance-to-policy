#!/bin/bash

pushd deploy

git fetch
diff=`git diff main origin/main --name-only  --relative=deployments | wc -l | tr -d '[:space:]'`
if [[ "$diff" == "0" ]];then
  echo Nothing to do since no diff
else
  echo Update deployments
  git merge origin/main
  kubectl replace -f ./deployments/manifests.yaml
fi

popd