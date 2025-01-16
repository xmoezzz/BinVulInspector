from collections import Counter
import os, sys, json, binascii
from glibc_symbols import is_glibc_symbol

def process_instr(instr: str):
    name, args = instr.split(" ")
    args = args.split(",")

    for i in range(len(args)):
        arg = args[i]

        if arg.startswith("["):
            arg = arg[1:-1]

        if arg.startswith("call:"):
            callee = arg[len("call:") :]
            if is_glibc_symbol(callee):
                arg = "libc:{}".format(callee)
            else:
                arg = "call:unknown"

        args[i] = arg

    return "{}_{}".format(name.lower(), "_".join(args))


def load_pcode(path: str, ignore_unnamed: bool):
    with open(path, "rb") as f:
        obj = json.load(f)

    ret = []
    for sub in obj["program"]["term"]["subs"]:
        term = sub["term"]
        if not "ops" in term:
            continue

        name = term["name"]
        if ignore_unnamed and name.startswith("FUN_"):
            continue

        uuid = term.get("uuid", None)

        ret.append(
            (
                name,
                uuid,
                int(sub["tid"]["address"], 16),
                [process_instr(op) for op in term["ops"]],
            )
        )

    return ret


def generate_learning_data(json_path: str, ignore_unnamed: bool):
    import unit
    
    assert json_path.endswith(".json")

    BS = unit.BinarySummary()

    file_name = os.path.basename(json_path)[: -len(".json")]
    _, arch, compiler, optlevel = file_name.rsplit("-", 3)

    BS.dir_name = os.path.dirname(json_path)
    BS.bin_name = file_name
    BS.arch = arch
    BS.compiler = compiler
    BS.opt_level = optlevel

    label = "{} {}\n".format(compiler, optlevel)

    with open(json_path, "rb") as f:
        obj = json.load(f)

    for sub in obj["program"]["term"]["subs"]:
        BSFS = unit.FunctionSummary()

        term = sub["term"]

        if not "ops" in term:
            continue

        name = term["name"]
        if ignore_unnamed and name.startswith("FUN_"):
            continue
        BSFS.fn_name = name
        BSFS.fn_start = int(sub["tid"]["address"], 16)

        for op in term["ops"]:
            instr = process_instr(op)
            BSFS.normalized_instrs.append(instr)

        BS.fns_summaries.append(BSFS)

    corpus_data = ""
    corpus_ctr = 0
    corpus_voca = Counter()

    for fs in BS.fns_summaries:
        normalized_instrs = fs.normalized_instrs
        corpus_data += "\t".join(
            [BS.bin_name, fs.fn_name, ",".join(normalized_instrs), label]
        )
        corpus_voca += Counter(normalized_instrs)
        corpus_ctr += 1

    return BS, corpus_ctr, corpus_data, corpus_voca


if __name__ == "__main__":
    input_dir = sys.argv[1]
    output_path = sys.argv[2]

    corpus = []

    with open(output_path, "w") as corpus_out:
        for fp in os.listdir(input_dir):
            if not fp.endswith(".json"):
                continue

            fp = os.path.join(input_dir, fp)
            BS, corpus_ctr, corpus_data, corpus_voca = generate_learning_data(fp, False)
            corpus_out.write(corpus_data)
