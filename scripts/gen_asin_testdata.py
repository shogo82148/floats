#!/usr/bin/env python3
# Generates testdata/asin_128.txt or testdata/asin_256.txt.
# Each line contains the bits of x and the correctly rounded asin(x) in hexadecimal.
#
# Usage: python3 scripts/gen_asin_testdata.py [128|256]

import random
import sys
import mpmath

BITS = int(sys.argv[1]) if len(sys.argv) > 1 else 128
P, EB = {128: (112, 15), 256: (236, 19)}[BITS]
N = {128: 64, 256: 256}[BITS]  # the number of the table intervals of atan
B = (1 << (EB - 1)) - 1
width = (P + EB + 1) // 4
rnd = random.Random(BITS)
WORK = 4 * P + 64


def enc(m, e, s=0):
    return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))


def dec(v):
    s = v >> (P + EB)
    e = (v >> P) & ((1 << EB) - 1)
    if e == 0:
        x = mpmath.mpf(v & ((1 << P) - 1)) * mpmath.mpf(2) ** (1 - B - P)
    else:
        x = mpmath.mpf((v & ((1 << P) - 1)) | (1 << P)) * mpmath.mpf(2) ** (e - B - P)
    return -x if s else x


def round_bits(x):
    s = 1 if x < 0 else 0
    x = abs(x)
    if x == 0:
        return s << (P + EB)
    e = max(int(mpmath.floor(mpmath.log(x, 2))), 1 - B)
    m = int(mpmath.nint(x * mpmath.mpf(2) ** (P - e)))
    if m >> (P + 1):
        m >>= 1
        e += 1
    if m >> P == 0:
        return m | (s << (P + EB))
    return enc(m, e, s)


def random_normal(emin, emax, s=None):
    if s is None:
        s = rnd.getrandbits(1)
    return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)


def near(value, s):
    with mpmath.workprec(WORK):
        return round_bits(value) | (s << (P + EB))


inputs = []
# random values in [-1, 1], across the exponent range
for _ in range(1000):
    inputs.append(random_normal(-(B - 1), -1))
for _ in range(1500):
    inputs.append(random_normal(-12, -1))
# random subnormal values, including the smallest and largest
for _ in range(100):
    inputs.append((rnd.getrandbits(P) or 1) | (rnd.getrandbits(1) << (P + EB)))
inputs += [1, (1 << P) - 1, (1 << (P + EB)) | 1]
# exactly +-1, the neighbours of 1, and the smallest normal value
for s in (0, 1):
    inputs += [enc(1 << P, 0, s), enc((2 << P) - 1, -1, s), enc((2 << P) - 2, -1, s), enc(1 << P, 1 - B, s)]
# close to 1 from below, where sqrt(1-x**2) is tiny
for d in range(1, 60):
    for s in (0, 1):
        inputs.append(enc((2 << P) - 1 - d, -1, s))
for _ in range(300):
    d = rnd.getrandbits(rnd.randint(1, P))
    for s in (0, 1):
        inputs.append(enc((2 << P) - 1 - (d % (1 << P)), -1, s))
# x = sin(atan(t)) for t near the table breakpoints i/N and (i+1/2)/N
for i in range(1, N + 1):
    for off in (0, 0.5):
        for d in (-1e-30, 0, 1e-30):
            t = mpmath.mpf(i + off) / N * (1 + d)
            for u in (t, 1 / t):
                with mpmath.workprec(WORK):
                    v = mpmath.sin(mpmath.atan(u))
                for s in (0, 1):
                    inputs.append(near(v, s))
# x where the series is used directly, and near 1/sqrt(2)
for d in (-1e-30, 0, 1e-30):
    for s in (0, 1):
        with mpmath.workprec(WORK):
            inputs.append(near(mpmath.sin(mpmath.atan(mpmath.mpf(1) / (2 * N))) * (1 + d), s))
            inputs.append(near(mpmath.sqrt(mpmath.mpf(1) / 2) * (1 + d), s))
# tiny values, where asin(x) = x up to the rounding
for k in range(1, 400 if BITS == 128 else 800, 3 if BITS == 128 else 7):
    for s in (0, 1):
        inputs.append(near(mpmath.mpf(2) ** (-k) * (1 + rnd.random()), s))
# powers of two
for k in range(-30, 0):
    for s in (0, 1):
        inputs.append(enc(1 << P, k, s))

with open(f"testdata/asin_{BITS}.txt", "w") as f:
    for v in inputs:
        with mpmath.workprec(WORK):
            r = round_bits(mpmath.asin(dec(v)))
        f.write(f"{v:0{width}x} {r:0{width}x}\n")
