#!/bin/bash
kubectl  -n vault  delete  pvc  --all
kubectl -n vault get pv | grep vault | awk '{print $1}' | xargs kubectl delete pv
