#!/usr/bin/env python3
# Generates testdata/erfc256_kernels.txt, the reference values of the kernels of Float256 Erfc.
# Each line contains the kind of the kernel (S for erf256Scaled, L for erfc256Large), the argument x as Float256 in
# hexadecimal, the mantissa of erfc(x) in [1, 2) in fixed point with 320 fractional bits in 384-bit hexadecimal,
# and the exponent of erfc(x).
#
# Usage: python3 scripts/gen_erfc256_kernel_testdata.py

import random

import mpmath

from gen_gamma_testdata import round_bits, to_mpf

mpmath.mp.prec = 1500
F = 320
P, EB = 236, 19


def main():
    rnd = random.Random(2563)
    lines = []

    def add(kind, x):
        bits = round_bits(x, P, EB)
        xv = to_mpf(bits, P, EB)
        y = mpmath.erfc(xv)
        e = int(mpmath.floor(mpmath.log(y, 2)))
        m = int(mpmath.floor(y / mpmath.mpf(2) ** e * mpmath.mpf(2) ** F))
        lines.append(f"{kind} {bits:064x} {m:096x} {e}")

    def frac():
        # a random number in [0, 1) with more bits than Float256
        return mpmath.mpf(rnd.getrandbits(300)) / mpmath.mpf(2) ** 300

    for k in range(64, 256):
        for _ in range(3):
            add("S", (mpmath.mpf(k) + frac()) / 16)
    for _ in range(200):
        add("L", mpmath.mpf(16) + frac() * 48)
    for _ in range(200):
        add("L", mpmath.mpf(64) + frac() * 362)
    for t in (16, 17, 32, 64, 100, 256, 426):
        add("L", mpmath.mpf(t))
    with open("testdata/erfc256_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
