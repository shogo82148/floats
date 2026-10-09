#!/usr/bin/env python3
# Generates testdata/y0_128.txt.
# Each line contains the bits of x and the correctly rounded Y0(x) in hexadecimal.
#
# Usage: python3 scripts/gen_y0_128_testdata.py

import random

import mpmath

P, EB = 112, 15
B = (1 << (EB - 1)) - 1
mpmath.mp.prec = 600
rnd = random.Random(1280)


def enc(x):
    s = 1 if x < 0 else 0
    x = abs(x)
    if x == 0:
        return s << (P + EB)
    e = int(mpmath.floor(mpmath.log(x, 2)))
    m = int(mpmath.nint(x * mpmath.mpf(2) ** (P - e)))
    if m >= 1 << (P + 1):
        m >>= 1
        e += 1
    assert 1 - B <= e <= B
    return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))


def dec(v):
    e = (v >> P) & ((1 << EB) - 1)
    return mpmath.mpf((v & ((1 << P) - 1)) | (1 << P)) * mpmath.mpf(2) ** (e - B - P)


def main():
    xs = []
    # tiny arguments
    for e in (-4000, -1000, -100, -40, -20):
        xs.append(mpmath.mpf(1 + rnd.random()) * mpmath.mpf(10) ** e)
    # the power series, the Taylor series and the asymptotic expansion
    for lo, hi, n in ((1e-6, 2, 150), (2, 8, 150), (8, 64, 200), (64, 1000, 100), (1000, 1e6, 50)):
        for _ in range(n):
            xs.append(mpmath.mpf(lo) * mpmath.mpf(hi / lo) ** mpmath.mpf(rnd.random()))
    # around the boundaries and the zeros
    for b in (2, 8, 64):
        for d in (-1, 1):
            xs.append(dec(enc(mpmath.mpf(b)) + d))  # the adjacent representable values
    for k in range(1, 40):
        z = mpmath.findroot(lambda t: mpmath.bessely(0, t), (k - 0.75) * mpmath.pi, tol=mpmath.mpf(10) ** -150)
        xs.append(z * (1 + mpmath.mpf(rnd.random() - 0.5) * 1e-6))
    # huge arguments
    for e in (30, 300, 3000, 4900):
        for _ in range(3):
            xs.append(mpmath.mpf(1 + rnd.random()) * mpmath.mpf(10) ** e)

    lines = []
    for x in xs:
        xb = enc(x)
        xr = dec(xb)
        lines.append(f"{xb:032x} {enc(mpmath.bessely(0, xr)):032x}")
    with open("testdata/y0_128.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
