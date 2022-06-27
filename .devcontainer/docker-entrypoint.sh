#!/bin/sh

# link data assets for local run and test.
sudo mkdir /data
sudo ln -s /workspaces/catalog/pkg/components /data/components
sudo ln -s /workspaces/catalog/pkg/catalog.yaml /data/config.yaml