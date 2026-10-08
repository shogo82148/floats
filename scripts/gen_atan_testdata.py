#!/usr/bin/env python3
# Generates testdata/atan_128.txt or testdata/atan_256.txt.
# Each line contains the bits of x and the correctly rounded atan(x) in hexadecimal.
#
# Usage: python3 scripts/gen_atan_testdata.py [128|256]

import random
import sys
import mpmath

BITS = int(sys.argv[1]) if len(sys.argv) > 1 else 128
P, EB = {128: (112, 15), 256: (236, 19)}[BITS]
N = {128: 64, 256: 256}[BITS]  # the number of the table intervals
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
    if e > B:
        return (s << (P + EB)) | (((1 << EB) - 1) << P)
    if m >> P == 0:
        return m | (s << (P + EB))
    return enc(m, e, s)


def random_normal(emin, emax, s=None):
    if s is None:
        s = rnd.getrandbits(1)
    return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)


def near(value, s):
    # a Float with the value close to the given positive value.
    with mpmath.workprec(WORK):
        return round_bits(value) | (s << (P + EB))


inputs = []
# random values of both signs, across the whole exponent range
for _ in range(1000):
    inputs.append(random_normal(-(B - 1), B))
# random values around 1, where the argument is reduced with the table
for _ in range(2000):
    inputs.append(random_normal(-12, 12))
# random subnormal values, including the smallest and largest
for _ in range(100):
    inputs.append((rnd.getrandbits(P) or 1) | (rnd.getrandbits(1) << (P + EB)))
inputs += [1, (1 << P) - 1, (1 << (P + EB)) | 1]
# the largest and the smallest normal values
for s in (0, 1):
    inputs += [enc((2 << P) - 1, B, s), enc(1 << P, 1 - B, s)]
# near the table breakpoints i/N and (i+1/2)/N, and their reciprocals
for i in range(1, N + 1):
    for off in (0, 0.5):
        for d in (-1e-30, 0, 1e-30):
            v = mpmath.mpf(i + off) / N * (1 + d)
            for s in (0, 1):
                inputs.append(near(v, s))
                inputs.append(near(1 / v, s))
# 1/N/2 where the series is used directly, and 1
for d in (-1e-30, 0, 1e-30):
    for s in (0, 1):
        inputs.append(near(mpmath.mpf(1) / (2 * N) * (1 + d), s))
        inputs.append(near(mpmath.mpf(2) * N * (1 + d), s))
        inputs.append(near(1 + d, s))
# tiny and huge values, where atan(x) = x or pi/2 up to the rounding
for k in range(1, 400 if BITS == 128 else 800, 3 if BITS == 128 else 7):
    for s in (0, 1):
        inputs.append(near(mpmath.mpf(2) ** (-k) * (1 + rnd.random()), s))
        inputs.append(near(mpmath.mpf(2) ** k * (1 + rnd.random()), s))
# powers of two
for k in range(-30, 31):
    for s in (0, 1):
        inputs.append(enc(1 << P, k, s))

with open(f"testdata/atan_{BITS}.txt", "w") as f:
    for v in inputs:
        with mpmath.workprec(WORK):
            r = round_bits(mpmath.atan(dec(v)))
        f.write(f"{v:0{width}x} {r:0{width}x}\n")
