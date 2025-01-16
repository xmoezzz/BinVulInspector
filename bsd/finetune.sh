#!/bin/sh
python3 binshot.py \
    --bert_model_path models/pretrain_combined_4/model_bert/bert_ep1.model \
    --vocab_path /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/pretrain.combined.corpus.voca \
    --output_path models/downstream_combined_4 \
    --result_path result \
    --train_dataset /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/binsim.combined.train.corpus.txt \
    --valid_dataset /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/binsim.combined.valid.corpus.txt \
    --test_dataset /media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/binsim.combined.test.corpus.txt