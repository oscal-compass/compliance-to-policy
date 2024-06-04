#!/bin/bash

component_definition=./plugins_public/tests/data/heterogeneous/component-definition.json
out=./plugins_public/tests/data/heterogeneous

while [[ $# -gt 0 ]]; do
  case $1 in
    -c|--component_definition)
      component_definition="$2"
      shift
      shift
      ;;
    -o|--out)
      out="$2"
      shift
      shift
      ;;
    -*|--*)
      echo "Unknown option $1"
      exit 1
      ;;
    *)
      POSITIONAL_ARGS+=("$1") # save positional arg
      shift # past argument
      ;;
  esac
done

mkdir -p $out/auditree
python samples_public/auditree/compliance_to_policy.py -c $component_definition -o $out/auditree/auditree.json
python samples_public/kyverno/compliance_to_policy.py -c $component_definition -o $out/kyverno
python samples_public/ocm/compliance_to_policy.py -c $component_definition -o $out/ocm