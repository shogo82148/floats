#!/usr/bin/env python3
# Generates testdata/erf256_kernels.txt, the reference values of the kernel of Float256 Erf.
# Each line contains the argument x in fixed point with 320 fractional bits in 384-bit hexadecimal,
# the result in the same format, 0 or 1 which is the same as the scaled of the kernel, and the exponent of the result.
# If the flag is 0, the result is erf(x) and the exponent is 0. Otherwise, the result is the mantissa m in [1, 2) of
# erfc(x) * 2**237 = m * 2**e.
#
# Usage: python3 scripts/gen_erf256_kernel_testdata.py

import random

import mpmath

mpmath.mp.prec = 1500
F = 320
LARGE_CELL = 112
CELLS = 204


def fix(x):
    return int(mpmath.floor(x * mpmath.mpf(2) ** F))


def main():
    rnd = random.Random(2562)
    lines = []
    for k in range(0, CELLS):
        for _ in range(3):
            x = (mpmath.mpf(k) + mpmath.mpf(rnd.random())) / 16
            if k == 0:
                x = mpmath.mpf(2) ** -8 + x
            xf = fix(x)
            x = mpmath.mpf(xf) / mpmath.mpf(2) ** F
            if k >= LARGE_CELL:
                y = mpmath.erfc(x) * mpmath.mpf(2) ** 237
                e = int(mpmath.floor(mpmath.log(y, 2)))
                lines.append(f"{xf:096x} {fix(y / mpmath.mpf(2) ** e):096x} 1 {e}")
            else:
                lines.append(f"{xf:096x} {fix(mpmath.erf(x)):096x} 0 0")
    with open("testdata/erf256_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
