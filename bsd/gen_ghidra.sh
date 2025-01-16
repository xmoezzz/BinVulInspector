#!/bin/bash

dir=$1
worker_path=$(realpath ./gen_ghidra_worker.sh)
temp_file=$(mktemp)

echo "Writing command list to $temp_file"

for file in "$dir/"*
  do
    file_full_path=$(realpath $file)
    json_full_path=${file_full_path}.json

    if [ -e $json_full_path ]
    then
      echo "Skipping $file_full_path"
    else
      if [[ -x $file ]]
      then
        echo "$worker_path $file_full_path $json_full_path" >> $temp_file
      else
        echo "Skipping $file_full_path"
      fi
    fi
done

parallel -j 80 < $temp_file
