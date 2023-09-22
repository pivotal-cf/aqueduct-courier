#!/bin/bash

cd aqueduct-courier

grype . --scope AllLayers --add-cpes-if-none --fail-on "negligible" -vv
