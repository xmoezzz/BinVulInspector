glibc_symbols = set()

with open("glibc_symbols.txt", 'r') as f:
    for line in f:
        glibc_symbols.add(line.strip())

def is_glibc_symbol(name: str):
    return name in glibc_symbols
