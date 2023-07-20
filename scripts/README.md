## 
1. Set environment variables
    ```
    export GITHUB_ORG=
    export GITHUB_REPO=
    export GITHUB_USER=
    export GITHUB_TOKEN=
    export KUBECONFIG=
    ```
1. Init directory
    ```
    init.sh
    ```
1. Collect results
    ```
    collect.sh
    ```

# Install collect cronjob
1. Create github token in secret
```
kubectl -n c2p create secret generic --save-config collect-ocm-status-secret --from-literal=github-token=$GITHUB_TOKEN
```
1. Deploy c2p status collector
```
kubectl apply -f ./scripts/collect
```

# Install GitOps
1. Create Secret
```
kubectl -n c2p create secret generic --save-config git-secret --from-literal=user=$GITHUB_USER --from-literal=accessToken=$GITHUB_TOKEN
```
1. Deploy channel and subscription
```
kubectl -n c2p apply -f ./scripts/gitops
```

