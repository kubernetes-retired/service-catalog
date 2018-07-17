#!/bin/bash
if [[ $TRAVIS == "true" ]]; then
  docker build -f Dockerfile.controller -t docker.io/service-catalog/podpreset-crd:latest .
else
  # Dockerfile is using multi-stage build and I'm not messing with upgrading Docker on my host
  # unfortunately, imagebuilder needs to have https://github.com/openshift/imagebuilder/pull/74
  imagebuilder -f Dockerfile.controller -t docker.io/service-catalog/podpreset-crd:latest .
fi
#kubebuilder create config --controller-image docker.io/service-catalog/podpreset-crd:latest --name podpreset-crd
