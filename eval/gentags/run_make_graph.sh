#!/bin/bash

set -exuo pipefail

python ./make_graph_data.py ./logs ./results/graph-data.json
python ./make_graph.py ./results/graph-data.json ./results/graph-gentags.svg
