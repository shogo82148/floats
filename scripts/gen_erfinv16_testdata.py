#!/usr/bin/env python3
# Generates testdata/erfinv16.txt, the correctly rounded erfinv(x) for all the positive Float16 values x < 1.
# Each line contains the bits of x and the bits of the result in hexadecimal.
#
# Usage: python3 scripts/gen_erfinv16_testdata.py

import mpmath

from gen_gamma_testdata import round_bits, to_mpf

mpmath.mp.prec = 200

with open("testdata/erfinv16.txt", "w") as f:
    for b in range(0, 0x3C00):
        y = mpmath.erfinv(to_mpf(b, 10, 5))
        f.write(f"{b:04x} {round_bits(y, 10, 5):04x}\n")
