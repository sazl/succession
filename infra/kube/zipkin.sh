#!/bin/bash
kubectl.exe create -f zipkin.pvc.yaml
kubectl.exe create -f http://repo1.maven.org/maven2/io/fabric8/zipkin/zipkin-starter/0.1.12/zipkin-starter-0.1.12-kubernetes.yml
