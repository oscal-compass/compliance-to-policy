#!/bin/bash

interval=15s

while getopts :-: opt; do
    optarg="$OPTARG"
    if [[ "$opt" = - ]]; then
        opt="-${OPTARG%%=*}"
        optarg="${OPTARG/${OPTARG%%=*}/}"
        optarg="${optarg#=}"

        if [[ -z "$optarg" ]] && [[ ! "${!OPTIND}" = -* ]]; then
            optarg="${!OPTIND}"
            shift
        fi
    fi

    case "-$opt" in
        --interval)
            interval="$optarg"
            ;;
        --)
            break
            ;;
        -\?)
            exit 1
            ;;
        --*)
            echo "$0: illegal option -- ${opt##-}" >&2
            exit 1
            ;;
    esac
done
shift $((OPTIND - 1))

kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# wait for install
count=0
while [[ $count -lt 3 ]]
do
  kubectl wait --for=condition=Ready pods --all -n argocd
  if [[ "$?" != 0 ]];then
    echo retry in 5s
    sleep 5
    let count=count+1
  fi
  break
done

# # create admin password
kubectl port-forward svc/argocd-server -n argocd 8080:443 & pid=$!; sleep 3
argocd admin initial-password -n argocd
admin_pass=$(kubectl -n argocd get secret argocd-initial-admin-secret -o=jsonpath='{.data.password}' | base64 -d)
kill $pid

# setup argocd
kubectl config set-context --current --namespace=argocd
kubectl -n argocd patch configmap argocd-cm --type merge -p '{"data":{"timeout.reconciliation":"'$interval'"}}'
kubectl -n argocd rollout restart deploy argocd-repo-server
kubectl -n argocd rollout restart sts argocd-application-controller
kubectl wait --for=condition=Ready pods --all -n argocd