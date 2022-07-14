#!/bin/sh

# link data assets for local run and test.
sudo mkdir /data
sudo mkdir -p /var/run/secrets/kubernetes.io/serviceaccount
sudo chown $(id -u):$(id -g) /var/run/secrets/kubernetes.io/serviceaccount
sudo ln -s $PWD/pkg/components /data/components
sudo ln -s $PWD/pkg/catalog.yaml /data/config.yaml