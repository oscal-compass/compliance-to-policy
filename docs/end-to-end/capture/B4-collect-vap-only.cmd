python docs/end-to-end/reference/collect_and_assess.py \
    --context kind-storefront \
    -c docs/end-to-end/oscal/component-definitions/kubecon-storefront/component-definition.json \
    --catalog docs/end-to-end/oscal/catalogs/nist_800_53_resolved/catalog.json \
    -e vap-cel -o out-vap
