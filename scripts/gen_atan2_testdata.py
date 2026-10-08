#!/usr/bin/env python3
# Generates testdata/atan2_128.txt.
# Each line contains the bits of y, the bits of x, and the correctly rounded atan2(y, x) in hexadecimal.
#
# Usage: python3 scripts/gen_atan2_testdata.py

import random
import mpmath

P, EB = 112, 15
B = (1 << (EB - 1)) - 1
width = (P + EB + 1) // 4
rnd = random.Random(128)
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


def random_subnormal(s=None):
    if s is None:
        s = rnd.getrandbits(1)
    frac = rnd.getrandbits(P) or 1
    return frac | (s << (P + EB))


def ratio_input(ratio, ex, sy, sx):
    # (y, x) with y/x close to ratio, |x| ~ 2**ex.
    with mpmath.workprec(WORK):
        xm = enc(rnd.getrandbits(P) | (1 << P), ex, sx)
        ym = round_bits(abs(dec(xm)) * ratio)
    return ym | (sy << (P + EB)), xm


pairs = []
# random pairs of similar magnitude, all sign combinations
for _ in range(1500):
    ex = rnd.randint(-200, 200)
    pairs.append((random_normal(ex - 3, ex + 3), random_normal(ex - 3, ex + 3)))
# random pairs across the whole exponent range
for _ in range(500):
    pairs.append((random_normal(-(B - 1), B), random_normal(-(B - 1), B)))
# random pairs where one of them is subnormal
for _ in range(150):
    pairs.append((random_subnormal(), random_normal(-(B - 1), B)))
    pairs.append((random_normal(-(B - 1), B), random_subnormal()))
    pairs.append((random_subnormal(), random_subnormal()))
# ratios just around the table breakpoints (i+1/2)/64 and the points i/64
for i in range(0, 65):
    for off in (0, 0.5):
        for d in (-1e-30, 0, 1e-30):
            ratio = mpmath.mpf(i + off) / 64 * (1 + d)
            if ratio == 0:
                continue
            for sy, sx in ((0, 0), (1, 0), (0, 1), (1, 1)):
                pairs.append(ratio_input(ratio, rnd.randint(-50, 50), sy, sx))
                # the reciprocal ratio, where the arguments are swapped
                pairs.append(ratio_input(1 / ratio, rnd.randint(-50, 50), sy, sx))
# tiny and huge ratios
for k in range(1, 400, 3):
    for sy, sx in ((0, 0), (1, 1)):
        pairs.append(ratio_input(mpmath.mpf(2) ** (-k), rnd.randint(-100, 100), sy, sx))
        pairs.append(ratio_input(mpmath.mpf(2) ** k, rnd.randint(-100, 100), sy, sx))
# extreme exponents
for ey, ex in ((-16494, 16383), (16383, -16494), (16383, 16383), (-16494, -16494), (-16382, 16383), (16383, -16382)):
    for sy, sx in ((0, 0), (0, 1), (1, 0), (1, 1)):
        for _ in range(3):
            def mk(e, s):
                if e < 1 - B:
                    return (rnd.getrandbits(P) or 1) | (s << (P + EB))
                return random_normal(e, e, s)
            pairs.append((mk(ey, sy), mk(ex, sx)))
# y == x and y == -x, and powers of two
for _ in range(40):
    v = random_normal(-300, 300, 0)
    for sy, sx in ((0, 0), (1, 0), (0, 1), (1, 1)):
        pairs.append((v | (sy << (P + EB)), v | (sx << (P + EB))))
for k in range(-30, 31):
    for sy, sx in ((0, 0), (1, 1)):
        pairs.append((enc(1 << P, k, sy), enc(1 << P, 0, sx)))
        pairs.append((enc(1 << P, 0, sy), enc(1 << P, k, sx)))

with open("testdata/atan2_128.txt", "w") as f:
    for y, x in pairs:
        with mpmath.workprec(WORK):
            r = round_bits(mpmath.atan2(dec(y), dec(x)))
        f.write(f"{y:0{width}x} {x:0{width}x} {r:0{width}x}\n")
