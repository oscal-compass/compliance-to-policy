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

