1. Install Kyverno
    ```
    kubectl create -f https://github.com/kyverno/kyverno/releases/download/v1.10.0/install.yaml
    ```
1. Install ArgoCD
    ```
    ./scripts/install-argocd.sh --interval 60s
    ```
1. Setup ArgoCD
    ```
    ./scripts/setup-argocd.sh --user yana1205 --token <PAT> --org yana1205 --repo c2p-for-kyverno1008-config --path kyverno-policies --appname c2p
    ```

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
kubectl -n c2p create secret generic --save-config kyverno-policy-report-secret --from-literal=token=$GITHUB_TOKEN --from-literal=user=$GITHUB_USER --from-literal=repo=$GITHUB_REPO --from-literal=org=$GITHUB_ORG
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

