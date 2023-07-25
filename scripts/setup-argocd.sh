#!/bin/bash

ns=default
appname=guestbook

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
        --user)
            user="$optarg"
            ;;
        --token)
            token="$optarg"
            ;;
        --org)
            org="$optarg"
            ;;
        --repo)
            repo="$optarg"
            ;;
        --path)
            path="$optarg"
            ;;
        --dest-ns)
            ns="$optarg"
            ;;
        --appname)
            appname="$optarg"
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

# setup argocd
admin_pass=$(kubectl -n argocd get secret argocd-initial-admin-secret -o=jsonpath='{.data.password}' | base64 -d)
kubectl config set-context --current --namespace=argocd
kubectl port-forward svc/argocd-server -n argocd 8080:443 & pid=$!; sleep 3
argocd login localhost:8080 --username=admin --password="${admin_pass}" --insecure
argocd repocreds add --upsert https://github.com/$org/$repo --username $user --password $token
# argocd app create $appname --repo https://github.com/$org/$repo.git --path $path --dest-server https://kubernetes.default.svc --dest-namespace $ns --sync-option Replace=true --sync-policy automated --allow-empty --auto-prune
argocd app create $appname --repo https://github.com/$org/$repo.git --path $path --dest-server https://kubernetes.default.svc --dest-namespace $ns --sync-option Replace=true --sync-policy automated --allow-empty
argocd app get $appname

kill $pid