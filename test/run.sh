#!/bin/bash

curl -k -X POST -H "Content-Type: application/json" -d @./AdmissionReviewPod.json https://localhost:9443/mutate--v1-pod \
| jq -r '.response.patch' \
| base64 -d \
| jq .

curl -k -X POST -H "Content-Type: application/json" -d @./AdmissionReviewIngress.json https://localhost:9443/mutate-networking-k8s-io-v1-ingress \
| jq -r '.response.patch' \
| base64 -d \
| jq .
