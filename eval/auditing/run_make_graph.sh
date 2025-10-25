#!/bin/bash

set -exuo pipefail

python ./make_graph.py ./results/genproof_fixed-block-num.json ./results/graph-auditing.svg
