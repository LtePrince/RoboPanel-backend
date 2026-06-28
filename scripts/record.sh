#!/bin/bash
set -e

source /home/xiaoqiang/miniconda3/etc/profile.d/conda.sh
conda activate openteach
source /home/xiaoqiang/amper/catkin_ws_vr/devel/setup.bash

cd /home/xiaoqiang/Desktop/Open-Teach
exec python data_collect.py "$@"
