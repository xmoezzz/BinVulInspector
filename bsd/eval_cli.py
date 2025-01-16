from voca import WordVocab
from binshot import SimilarityModel, SimDataset, collate_sim
from eval import ModelRunner
from normalize_ghidra import load_pcode
import itertools, argparse, os, heapq, json
from database import BinaryDatabase

import torch
import torch.nn as nn

BATCH_SIZE = 32


def batched(iterable, n):
    "Batch data into lists of length n. The last batch may be shorter."
    # batched('ABCDEFG', 3) --> ABC DEF G
    it = iter(iterable)
    while True:
        batch = list(itertools.islice(it, n))
        if not batch:
            return
        yield batch


if __name__ == "__main__":
    parser = argparse.ArgumentParser()

    parser.add_argument(
        "--vocab", type=str, required=True, help="Path to .corpus.voca file"
    )
    parser.add_argument(
        "--model", type=str, required=True, help="Path to sim_epn.model file"
    )
    parser.add_argument(
        "--db", type=str, required=True, help="Path to database directory"
    )
    parser.add_argument(
        "--input", type=str, required=True, help="Path to input Pcode JSON"
    )
    parser.add_argument("--output", type=str, required=True, help="Path to output JSON")

    args = parser.parse_args()
    with torch.no_grad():
        model = ModelRunner(args.vocab, args.model)
        db = BinaryDatabase(os.path.join(args.db, "functions.db"))

        pcode = load_pcode(args.input, False)

        db_info_batches = []
        db_emb_batches = []
        db_count = 0

    
        for batch in batched(db.iter_all(), BATCH_SIZE):
            info_batch = []
            text_batch = []

            for name, uuid, instrs in batch:
                info_batch.append((name, uuid))
                text_batch.append(instrs.split(","))

            db_info_batches.append(info_batch)
            db_emb_batches.append(model.generate_embeddings(text_batch))

            db_count += len(batch)

        print("[+] Loaded {} functions from DB".format(db_count))

        pc_info_batches = []
        pc_emb_batches = []
        pc_count = 0

        for batch in batched(pcode, BATCH_SIZE):
            info_batch = []
            text_batch = []

            for name, uuid, addr, instrs in batch:
                info_batch.append((name, uuid, addr))
                text_batch.append(instrs)

            pc_info_batches.append(info_batch)
            pc_emb_batches.append(model.generate_embeddings(text_batch))

            pc_count += len(batch)

        print("[+] Loaded {} functions from JSON".format(pc_count))

        output_funcs = []

        for pc_batch_i, pc_emb_batch in enumerate(pc_emb_batches):
            info_batch = pc_info_batches[pc_batch_i]

            for i, pc_emb in enumerate(pc_emb_batch):
                name, uuid, addr = info_batch[i]

                pc_emb_stacked_full = torch.stack([pc_emb] * BATCH_SIZE)

                h = []

                for db_batch_i, db_emb_batch in enumerate(db_emb_batches):
                    db_info_batch = db_info_batches[db_batch_i]

                    b = pc_emb_stacked_full
                    if len(b) != len(db_emb_batch):
                        b = torch.stack([pc_emb] * len(db_emb_batch))

                    results = (
                        model.model.compare(b, db_emb_batch)[:, 0].cpu().detach().numpy()
                    )

                    for result, db_info in zip(results, db_info_batch):
                        heap_item = (1 - result, db_info)
                        heapq.heappush(h, heap_item)

                sorted_results = []
                while len(h) > 0:
                    result_rev, db_info = heapq.heappop(h)
                    sorted_results.append(
                        {"sim": -result_rev + 1, "name": db_info[0], "cve_uuid": db_info[1]}
                    )

                output_funcs.append(
                    {"addr": hex(addr)[2:], "fname": name, "results": sorted_results}
                )

        with open(args.output, "w") as f:
            json.dump({"funcs": output_funcs}, f, ensure_ascii=False, indent=4)

        print("[+] Output written to {}".format(args.output))
