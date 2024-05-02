#!/bin/bash

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

if [[ "$appname" == "" ]];then
  echo "Need --appname"
  exit 1
fi

kubectl config set-context --current --namespace=argocd
kubectl port-forward svc/argocd-server -n argocd 8080:443 & pid=$!; sleep 3
argocd app delete -y $appname
while [[ "`kubectl -n argocd get applications $appname`" != "" ]]
do
  echo wait for $appname to be deleted
  sleep 5
done
kill $pid