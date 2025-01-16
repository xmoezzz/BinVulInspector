import contextlib, argparse, os
import sqlite3, zlib, typing

from normalize_ghidra import load_pcode


class BinaryDatabase:
    def __init__(self, db_path: str):
        self.db = sqlite3.connect(db_path)
        self.db.row_factory = sqlite3.Row

        self.db.execute(
            """
            CREATE TABLE IF NOT EXISTS functions (
                id INTEGER PRIMARY KEY,
                uuid TEXT NOT NULL,
                name TEXT NOT NULL,
                instr BLOB NOT NULL
            );
            """
        )

        self.db.commit()

    def close(self):
        self.db.close()

    def insert_function(self, uuid: str, fn_name: str, instrs: str):
        instr_comp = zlib.compress(instrs.encode("utf-8"))
        self.db.execute(
            """
            INSERT INTO functions VALUES(NULL, ?, ?, ?)
            """,
            (uuid, fn_name, instr_comp),
        )
        self.db.commit()

    def insert_function_many(self, functions: typing.List[typing.Tuple[str, str, str]]):
        for i in range(len(functions)):
            instr = functions[i][2]
            functions[i][2] = zlib.compress(instr.encode("utf-8"))

        self.db.executemany(
            """
            INSERT INTO functions(name, uuid, instr) VALUES(?, ?, ?)
            """,
            functions,
        )
        self.db.commit()

    def iter_all(self):
        with contextlib.closing(self.db.cursor()) as cursor:
            for row in cursor.execute("SELECT name, uuid, instr FROM functions"):
                yield (row['name'], row['uuid'], zlib.decompress(row['instr']).decode("utf-8"))

if __name__ == "__main__":
    parser = argparse.ArgumentParser()    

    parser.add_argument("--db", type=str, required=True, help="Path to database directory")
    parser.add_argument("-i", type=str, required=True, help="Path to pcode JSON file")

    args = parser.parse_args()

    os.makedirs(args.db, exist_ok=True)

    db = BinaryDatabase(os.path.join(args.db, "functions.db"))

    pcode = load_pcode(args.i, True)
    functions = [[x[0], x[1], ",".join(x[3])] for x in pcode]
    db.insert_function_many(functions)
    db.close()