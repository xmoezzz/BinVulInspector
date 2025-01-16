#!/bin/sh
python3 bert_mlm.py \
    --corpus_dataset /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/pretrain.combined.corpus.txt \
    --vocab_path /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/pretrain.combined.corpus.voca \
    --output_path models/pretrain_combined_4
