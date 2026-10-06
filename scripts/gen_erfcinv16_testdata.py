#!/usr/bin/env python3
# Generates testdata/erfcinv16.txt, the correctly rounded erfcinv(x) for all the Float16 values 0 < x <= 1.
# Each line contains the bits of x and the bits of the result in hexadecimal.
#
# Usage: python3 scripts/gen_erfcinv16_testdata.py

import mpmath

from gen_gamma_testdata import round_bits, to_mpf

mpmath.mp.prec = 300

with open("testdata/erfcinv16.txt", "w") as f:
    for b in range(1, 0x3C01):
        x = to_mpf(b, 10, 5)
        y = mpmath.erfinv(1 - x)
        f.write(f"{b:04x} {round_bits(y, 10, 5) if y != 0 else 0:04x}\n")
