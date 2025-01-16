#!/bin/bash

echo "Processing ${1} to ${2}"

cd /home/xuzhihua/work/bin-checker/sfs/bin2

./pcodeextractor -i "${1}" -o "${2}"
