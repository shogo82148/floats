#!/usr/bin/env python3
# Generates testdata/erf128_kernels.txt, the reference values of the kernel of Float128 Erf.
# Each line contains the argument x in fixed point with 192 fractional bits in 256-bit hexadecimal,
# the result in the same format, and 0 if the result is erf(x) or 1 if it is erfc(x) * 2**113.
#
# Usage: python3 scripts/gen_erf128_kernel_testdata.py

import random

import mpmath

mpmath.mp.prec = 700
F = 192
LARGE_CELL = 112


def fix(x):
    return int(mpmath.floor(x * mpmath.mpf(2) ** F))


def main():
    rnd = random.Random(1282)
    lines = []
    for k in range(0, 140):
        for _ in range(3):
            x = (mpmath.mpf(k) + mpmath.mpf(rnd.random())) / 16
            if k == 0:
                x = mpmath.mpf(2) ** -8 + x
            xf = fix(x)
            x = mpmath.mpf(xf) / mpmath.mpf(2) ** F
            if k >= LARGE_CELL:
                y, flag = mpmath.erfc(x) * mpmath.mpf(2) ** 113, 1
            else:
                y, flag = mpmath.erf(x), 0
            lines.append(f"{xf:064x} {fix(y):064x} {flag}")
    with open("testdata/erf128_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
