#!/bin/sh

# link data assets for local run and test.
sudo mkdir /data
sudo ln -s $PWD/pkg/components /data/components
sudo ln -s $PWD/pkg/catalog.yaml /data/config.yaml