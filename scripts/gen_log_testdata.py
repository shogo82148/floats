#!/usr/bin/env python3
# Generates testdata/log{128,256}.txt, testdata/log2_{128,256}.txt, or testdata/log10_{128,256}.txt, or testdata/log1p_{128,256}.txt.
# Each line contains the bits of x and the correctly rounded log(x), log2(x), log10(x), or log1p(x) in hexadecimal.
#
# Usage: python3 scripts/gen_log_testdata.py [128|256] [log|log2|log10|log1p]

import random
import sys
import mpmath

BITS = int(sys.argv[1]) if len(sys.argv) > 1 else 128
FUNC = sys.argv[2] if len(sys.argv) > 2 else "log"
P, EB = {128: (112, 15), 256: (236, 19)}[BITS]
B = (1 << (EB - 1)) - 1
width = (P + EB + 1) // 4
rnd = random.Random(BITS)


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
    # round the positive value x to the nearest representable value.
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
        return (s << (P + EB)) | (((1 << EB) - 1) << P)  # infinity
    if m >> P == 0:
        return m | (s << (P + EB))  # subnormal
    return enc(m, e, s)


def random_normal(emin, emax):
    return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), 0)


def random_subnormal():
    frac = rnd.getrandbits(P)
    if frac == 0:
        frac = 1
    return frac


inputs = []
# random normal values across a wide exponent range
inputs += [random_normal(-(B - 383), B - 383) for _ in range(300)]
# random subnormal values, including the smallest and largest
inputs += [random_subnormal() for _ in range(80)]
inputs += [1, (1 << P) - 1]
# near each table bucket boundary (idx transition), both just below and
# just above, where the reduction switches breakpoints
for idx in range(0, 256, 4):
    base = enc((1 << P) + (idx << (P - 8)), 0)
    for d in (-1, 0, 1):
        inputs.append(base + d)
# near 1 from both sides, at varying distance, where Log must not round log(c)
# for the bucket (0 or 255) that borders a power of two: from above, a is just
# past 1 at exponent 0; from below, a is just under 1 at exponent -1.
for e in range(-P - 2, -1):
    for d in range(3):
        inputs += [enc((1 << P) + d, e, 0), enc((2 << P) - 1 - d, e - 1, 0)]
# exactly 1, and the smallest perturbations around it
inputs.append(enc(1 << P, 0, 0))
inputs.append(enc((1 << P) + 1, 0, 0))
inputs.append(enc((2 << P) - 1, -1, 0))
# near other powers of two, where exp changes and log(c) must combine with
# k*ln2 without losing precision
for k in range(-20, 21):
    for d in range(3):
        inputs += [enc((1 << P) + d, k, 0), enc((2 << P) - 1 - d, k - 1, 0)]
# near the smallest normal/largest subnormal boundary
inputs += [enc(1 << P, 1 - B, 0), (1 << P) - 1, enc((1 << P) + 1, 1 - B, 0)]

if FUNC == "log2":
    # exact powers of two, whose log2 is exact
    for k in range(-30, 31):
        inputs.append(enc(1 << P, k, 0))
    for _ in range(20):
        inputs.append(enc(1 << P, rnd.randint(-(B - 1), B), 0))

if FUNC == "log10":
    # powers of ten, whose log10 is an integer (the powers up to 10**48 are exact)
    for k in range(0, 49):
        with mpmath.workprec(4 * P + 64):
            inputs.append(round_bits(mpmath.mpf(10) ** k))
    for k in range(-40, 0):
        with mpmath.workprec(4 * P + 64):
            inputs.append(round_bits(mpmath.mpf(10) ** k))

if FUNC == "log1p":
    # log1p is defined for x > -1, and its inputs are concentrated near 0 and -1.
    inputs = []
    # random values of both signs across a wide exponent range
    for _ in range(200):
        inputs.append(random_normal(-(B - 383), B - 383))
    for _ in range(200):
        inputs.append(random_normal(-300, 300) | rnd.getrandbits(1) << (P + EB))
    # random negative values in (-1, 0)
    for _ in range(100):
        inputs.append(random_normal(-60, -1) | 1 << (P + EB))
    inputs += [random_subnormal() for _ in range(20)]
    inputs += [1, (1 << P) - 1, (1 << (P + EB)) | 1]
    # around 2**-8, where the series switches to the table-driven reduction
    for e in (-10, -9, -8, -7):
        for d in range(3):
            for s in (0, 1):
                inputs += [enc((1 << P) + d, e, s), enc((2 << P) - 1 - d, e, s)]
    # close to 0 at every exponent, from both sides
    for e in range(-P - 4, -7):
        for s in (0, 1):
            inputs += [enc((1 << P), e, s), enc((1 << P) + 1, e, s), enc((2 << P) - 1, e, s)]
    # close to -1 from above, where 1+x is tiny, and close to the bucket boundaries of 1+x
    for d in range(40):
        inputs.append(enc((2 << P) - 1 - d, -1, 1))
    for idx in range(0, 256, 8):
        base = enc((1 << P) + (idx << (P - 8)), 0)
        inputs += [base, base + 1]
        # x = -(1 - c) for c in the same bucket: 1+x = c/2
        inputs.append(enc((1 << P) + (idx << (P - 8)), -1, 1))
    # powers of two and their neighbours
    for k in range(-20, 21):
        for d in range(3):
            inputs += [enc((1 << P) + d, k, 0), enc((2 << P) - 1 - d, k - 1, 0)]
    inputs += [enc(1 << P, 1 - B, 0), enc((1 << P) + 1, 1 - B, 0)]
    # the largest finite value
    inputs.append(enc((2 << P) - 1, B, 0))
    with mpmath.workprec(4 * P + 64):
        inputs = [v for v in inputs if dec(v) > -1]

LOG = {
    "log": mpmath.log,
    "log2": lambda x: mpmath.log(x, 2),
    "log10": mpmath.log10,
    "log1p": mpmath.log1p,
}[FUNC]
name = f"testdata/{FUNC}_{BITS}.txt" if FUNC != "log" else f"testdata/log{BITS}.txt"
with open(name, "w") as f:
    for v in inputs:
        with mpmath.workprec(4 * P + 64):
            x = dec(v)
        with mpmath.workprec(4 * P + 64):
            y = round_bits(LOG(x))
        f.write(f"{v:0{width}x} {y:0{width}x}\n")
