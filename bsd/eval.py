import os, sys
import typing
import statistics
import argparse
import random
import tqdm
import traceback
from glob import glob

import numpy as np
import torch
import torch.nn as nn
from torch.utils.data import Dataset, DataLoader, random_split

import warnings
from sklearn.exceptions import UndefinedMetricWarning

warnings.filterwarnings(action="ignore", category=UndefinedMetricWarning)

from bert_mlm import (
    BERTDataset,
    Linear,
    MultiheadAttention,
    Attention,
    collate_mlm,
    pad1d,
)
from bert_mlm import Convolution, PositionWiseFeedForward
from bert_mlm import EncoderPrenet, BERTEncoder, BertAdam, optim4GPU

import hparams as hp
from util import compute_prediction_metric
from util import write_metrics, write_pred_results
from voca import WordVocab

from binshot import SimilarityModel, SimDataset, collate_sim


class EvalDataset(Dataset):
    def __init__(self, corpus_path, vocab: WordVocab, encoding="utf-8"):
        self.vocab = vocab
        self.num_data = 0
        self.corpus_path = corpus_path
        self.encoding = encoding
        self.corpus = []

        with open(corpus_path, "r", encoding=encoding) as f:
            for line in tqdm.tqdm(f, desc="[+] Loading Dataset", total=self.num_data):
                parts = line.split("\t")

                bin_name = parts[0]
                fn_name = parts[1]
                tokens = parts[2].split(",")

                self.corpus.append((bin_name, fn_name, tokens))

            self.num_data = len(self.corpus)
            print("[+] Number of actual dataset loaded: %d" % self.num_data)

    def __len__(self):
        return self.num_data

    def __getitem__(self, item):
        bin_name, fn_name, instructions = self.corpus[item]

        voca_ins = []
        for i, insn in enumerate(instructions):
            if i >= hp.enc_maxlen - 3:
                break
            idx = self.vocab.voca_idx(insn)
            voca_ins.append(idx)

        return (
            bin_name,
            fn_name,
            [self.vocab.sos_index] + voca_ins + [self.vocab.eos_index],
        )


def collate_eval(batch):
    input_lens = [len(x[2]) for x in batch]
    max_x_len = max(input_lens)

    # chars
    instrs_pad = [pad1d(x[2], max_x_len) for x in batch]
    instrs = np.stack(instrs_pad)

    # position
    position = [pad1d(range(1, len + 1), max_x_len) for len in input_lens]
    position = np.stack(position)

    instrs = torch.tensor(instrs).long()
    position = torch.tensor(position).long()

    return {
        "mlm_input": instrs,
        "info": [(x[0], x[1]) for x in batch],
        "input_position": position,
    }


class ModelRunner:
    def __init__(self, vocab_path: str, model_path: str, seed=99, cuda="0"):
        self.device_loc = "cuda:" + cuda if torch.cuda.is_available() else "cpu"
        self.device = torch.device(self.device_loc)
        torch.cuda.set_device(self.device)

        random.seed(seed)
        np.random.seed(seed)
        torch.manual_seed(seed)

        self.vocab = WordVocab.load_vocab(vocab_path)
        print("[+] Loaded %d vocas from %s" % (self.vocab.vocab_size, vocab_path))

        self.model: SimilarityModel = (
            torch.load(model_path, map_location=self.device_loc).to(self.device)
            if torch.cuda.device_count() == 1
            else torch.load(model_path).to(self.device)
        )
        self.model.eval()
        print("[+] Loaded model in evaluation mode")

    def generate_embeddings(self, batch: typing.List[typing.List[str]]):
        input_lens = []
        tokens = []

        for fn_instrs in batch:
            voca_ins = [self.vocab.sos_index]
            for i, insn in enumerate(fn_instrs):
                if i >= hp.enc_maxlen - 3:
                    break
                voca_ins.append(self.vocab.voca_idx(insn))
            voca_ins.append(self.vocab.eos_index)

            input_lens.append(len(voca_ins))
            tokens.append(voca_ins)

        max_x_len = max(input_lens)

        # chars
        instrs_pad = [pad1d(x, max_x_len) for x in tokens]
        instrs = np.stack(instrs_pad)

        # position
        positions = [pad1d(range(1, len + 1), max_x_len) for len in input_lens]
        positions = np.stack(positions)

        instrs = torch.tensor(instrs).long().to(self.device)
        positions = torch.tensor(positions).long().to(self.device)

        return self.model.to_emb(instrs, positions)


def main():
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--seed", type=int, default=99, help="random seed for initialization"
    )
    parser.add_argument("--cuda", type=str, default="0", help="cuda device")

    args = parser.parse_args()

    device_loc = "cuda:" + args.cuda if torch.cuda.is_available() else "cpu"
    device = torch.device(device_loc)
    torch.cuda.set_device(device)

    random.seed(args.seed)
    np.random.seed(args.seed)
    torch.manual_seed(args.seed)

    vocab_path = "/media/xmoe/storage/buildroot-elf-5arch/buildroot-elf-5arch/corpus/pretrain.combined.corpus.voca"
    wv = WordVocab.load_vocab(vocab_path)
    print("[+] Loaded %d vocas from %s" % (wv.vocab_size, vocab_path))

    ft_model_path = "/home/xuzhihua/work/binshot/models/downstream_combined_3/model_sim/sim_ep0.model"
    model: SimilarityModel = (
        torch.load(ft_model_path, map_location=device_loc).to(device)
        if torch.cuda.device_count() == 1
        else torch.load(ft_model_path).to(device)
    )
    model.eval()
    print("[+] Loaded FT model")

    # vulns_dataset = EvalDataset("/home/xuzhihua/work/binshot/binary/0224/vulns.txt", wv)
    # vulns_data_loader = DataLoader(
    #     vulns_dataset,
    #     batch_size=8,
    #     collate_fn=lambda batch: collate_eval(batch),
    #     shuffle=False,
    # )

    # targets_dataset = EvalDataset(
    #     "/home/xuzhihua/work/binshot/binary/0224/targets.txt", wv
    # )
    # targets_data_loader = DataLoader(
    #     targets_dataset,
    #     batch_size=1,
    #     collate_fn=lambda batch: collate_eval(batch),
    #     shuffle=False,
    # )

    # vulns_iter = tqdm.tqdm(
    #     enumerate(vulns_data_loader),
    #     total=len(vulns_data_loader),
    #     bar_format="{l_bar}{r_bar}",
    # )

    # # targets_iter = tqdm.tqdm(
    # #     enumerate(targets_data_loader),
    # #     total=len(targets_data_loader),
    # #     bar_format="{l_bar}{r_bar}",
    # # )

    # vulns = []

    # with torch.no_grad():
    #     vuln_fn_names = set()
    #     for i, data in vulns_iter:
    #         bert_output = model.to_emb(data["mlm_input"].to(device), data["input_position"].to(device))

    #         for (bin_name, fn_name), bert_output in zip(data['info'], bert_output):
    #             vulns.append(((bin_name, fn_name), bert_output))
    #             vuln_fn_names.add(fn_name)

    #     total_count = 0
    #     top10_count = 0
    #     top5_count = 0
    #     top1_count = 0

    #     for i, data in enumerate(targets_data_loader):
    #         bin_name, fn_name = data['info'][0]

    #         if not fn_name in vuln_fn_names:
    #             continue

    #         target_bert_output = model.to_emb(data["mlm_input"].to(device), data["input_position"].to(device))[0]

    #         results = []

    #         for (v_bin_name, v_fn_name), v_bert in vulns:
    #             result = model.compare(target_bert_output, v_bert)[0].item()
    #             results.append((v_bin_name, v_fn_name, result))

    #         results.sort(key=lambda x: x[2], reverse=True)

    #         for j, result in enumerate(results[:10]):
    #             _v_bin_name, v_fn_name, _sim = result
    #             if fn_name == v_fn_name:
    #                 top10_count += 1
    #                 if j < 5:
    #                     top5_count += 1
    #                 if j == 0:
    #                     top1_count += 1
    #                 break

    #         total_count += 1

    #         print(total_count, top10_count / total_count, top5_count / total_count, top1_count / total_count)


if __name__ == "__main__":
    main()
