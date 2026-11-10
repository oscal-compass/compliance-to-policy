c2p tools posture-compare \
    -b out/posture.json -p out-vap/posture.json \
    --baseline-label 'every engine' --posture-label 'vap-cel only' \
    --condition 'when only ValidatingAdmissionPolicy is available'
