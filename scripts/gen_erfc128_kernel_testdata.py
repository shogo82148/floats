#!/usr/bin/env python3
# Generates testdata/erfc128_kernels.txt, the reference values of the kernels of Float128 Erfc.
# Each line contains the kind of the kernel (S for erf128Scaled, L for erfc128Large), the argument x as Float128 in
# hexadecimal, the mantissa of erfc(x) in [1, 2) in fixed point with 192 fractional bits in 256-bit hexadecimal,
# and the exponent of erfc(x).
#
# Usage: python3 scripts/gen_erfc128_kernel_testdata.py

import random

import mpmath

from gen_gamma_testdata import round_bits

mpmath.mp.prec = 700
F = 192
P, EB = 112, 15


def main():
    rnd = random.Random(1283)
    lines = []

    def add(kind, x):
        bits = round_bits(x, P, EB)
        with mpmath.workprec(700):
            from gen_gamma_testdata import to_mpf

            xv = to_mpf(bits, P, EB)
            y = mpmath.erfc(xv)
            e = int(mpmath.floor(mpmath.log(y, 2)))
            m = int(mpmath.floor(y / mpmath.mpf(2) ** e * mpmath.mpf(2) ** F))
        lines.append(f"{kind} {bits:032x} {m:064x} {e}")

    for k in range(64, 256):
        for _ in range(3):
            add("S", (mpmath.mpf(k) + mpmath.mpf(rnd.random())) / 16)
    for _ in range(300):
        add("L", mpmath.mpf(16) + mpmath.mpf(rnd.random()) * 90)
    for t in (16, 17, 32, 64, 100, 105):
        add("L", mpmath.mpf(t))
    with open("testdata/erfc128_kernels.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
