#!/usr/bin/env python3
# Generates testdata/lgamma256_kernels.txt, the reference values of the kernels of Float256 Lgamma.
# Each line contains the kind of the kernel, the argument, and the result in fixed point with 320 fractional bits
# in 384-bit hexadecimal. The result of the kernels of the negative arguments has the sign "-" or "+" as the first character.
# stirling has the exponent of the argument and the result in decimal as the last fields.
#
# Usage: python3 scripts/gen_lgamma256_kernel_testdata.py

import random

import mpmath

mpmath.mp.prec = 1200
F = 320


def fix(x):
    return int(mpmath.floor(x * mpmath.mpf(2) ** F))


def hexfix(v):
    return f"{v:096x}"


def rand_fix(rnd, lo, hi):
    return int(lo * 2**F) + rnd.randrange(int((hi - lo) * 2**F))


def signed(v):
    return ("-" if v < 0 else "+") + hexfix(abs(v))


def main():
    rnd = random.Random(2562)
    lines = []

    def to_mpf(v):
        return mpmath.mpf(v) / mpmath.mpf(2) ** F

    for lo, hi, n in [(0.001, 2, 40), (2, 48, 40), (48, 2048, 30), (2048, 32000, 30), (31.9, 32.2, 10), (63.5, 64.5, 10), (1023, 1030, 10)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            y = mpmath.loggamma(to_mpf(v))
            if abs(y) > mpmath.mpf(2) ** -18:
                lines.append(f"pos {hexfix(v)} {signed(fix(y))}")
    for lo, hi, n in [(0.001, 3, 40), (3, 100, 30), (100, 32000, 30)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            x = to_mpf(v)
            if v % (1 << F) == 0:
                continue
            y = mpmath.log(mpmath.pi) - mpmath.log(abs(mpmath.sin(mpmath.pi * x))) - mpmath.loggamma(1 + x)
            if abs(y) > mpmath.mpf(2) ** -18:
                lines.append(f"neg {hexfix(v)} {signed(fix(y))}")
    # Lgamma(z) = ym 2**ye for z = zm 2**ze >= 2**15, zm and ym in [1, 2)
    for ze in (15, 16, 20, 40, 100, 300, 329, 1000):
        for _ in range(8):
            zm = rand_fix(rnd, 1, 2)
            z = to_mpf(zm) * mpmath.mpf(2) ** ze
            y = mpmath.loggamma(z)
            ye = int(mpmath.floor(mpmath.log(y, 2)))
            lines.append(f"stirling {hexfix(zm)} {ze} {hexfix(fix(y / mpmath.mpf(2) ** ye))} {ye}")

    with open("testdata/lgamma256_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
