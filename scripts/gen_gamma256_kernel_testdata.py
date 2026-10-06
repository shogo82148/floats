#!/usr/bin/env python3
# Generates testdata/gamma256_kernels.txt, the reference values of the kernels of Float256 Gamma.
# Each line contains the kind of the kernel, the argument, and the result in fixed point with 320 fractional bits
# in 384-bit hexadecimal. exp and gamma have the exponent of the result in decimal as the last field.
#
# Usage: python3 scripts/gen_gamma256_kernel_testdata.py

import random

import mpmath

mpmath.mp.prec = 1000
F = 320


def fix(x):
    return int(mpmath.floor(x * mpmath.mpf(2) ** F))


def hexfix(v):
    return f"{v:096x}"


def rand_fix(rnd, lo, hi):
    """returns a random fixed point number in [lo, hi) with random fractional bits."""
    return int(lo * 2**F) + rnd.randrange(int((hi - lo) * 2**F))


def main():
    rnd = random.Random(256)
    lines = []

    def to_mpf(v):
        return mpmath.mpf(v) / mpmath.mpf(2) ** F

    def mant_exp(x):
        e = int(mpmath.floor(mpmath.log(x, 2)))
        return fix(x / mpmath.mpf(2) ** e), e

    for lo, hi, n in [(1, 3, 30), (3, 100, 20), (100, 20368, 30)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            lines.append(f"log {hexfix(v)} {hexfix(fix(mpmath.log(to_mpf(v))))}")
    # powers of two, and the neighbors, where the logarithm is close to a multiple of log(2)
    for k in (0, 1, 5, 6, 10, 14):
        for d in (0, 1, -1, 1 << 200):
            if k == 0 and d < 0:
                continue
            v = (1 << (F + k)) + d
            lines.append(f"log {hexfix(v)} {hexfix(fix(mpmath.log(to_mpf(v))))}")
    for lo, hi, n in [(48, 64, 30), (64, 200, 20), (200, 2048, 20), (2048, 20368, 20)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            lines.append(f"lgamma {hexfix(v)} {hexfix(fix(mpmath.loggamma(to_mpf(v))))}")
    for lo, hi, n in [(0, 1, 20), (1, 300, 30), (300, 181000, 30)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            m, e = mant_exp(mpmath.exp(to_mpf(v)))
            lines.append(f"exp {hexfix(v)} {hexfix(m)} {e}")
    for lo, hi, n in [(0, 1e-3, 5), (1e-3, 0.25, 20), (0.25, 0.5, 30)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            x = to_mpf(v)
            lines.append(f"sinc {hexfix(v)} {hexfix(fix(mpmath.sin(mpmath.pi * x) / (mpmath.pi * x)))}")
    for _ in range(40):
        v = (1 << F) + rnd.randrange(1 << F)
        lines.append(f"recip {hexfix(v)} {hexfix(fix(1 / to_mpf(v)))}")
    for lo, hi, n in [(1e-12, 1, 30), (1, 48, 30), (48, 500, 20), (500, 20367, 20)]:
        for _ in range(n):
            v = rand_fix(rnd, lo, hi)
            m, e = mant_exp(mpmath.gamma(to_mpf(v)))
            lines.append(f"gamma {hexfix(v)} {hexfix(m)} {e}")

    with open("testdata/gamma256_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
