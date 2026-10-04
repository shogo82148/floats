#!/usr/bin/env python3
# Generates testdata/pow32.txt.
# Each line contains the bits of x, y, and the correctly rounded x**y in hexadecimal.
#
# Usage: python3 scripts/gen_pow_testdata.py

import random
import struct
import mpmath

mpmath.mp.prec = 400


def f32(v):
    return struct.unpack("<f", struct.pack("<I", v))[0]


def bits32(f):
    return struct.unpack("<I", struct.pack("<f", f))[0]


def round32(y):
    # round y to the nearest Float32 value, ties to even.
    if y == 0:
        return 0
    s = 0x80000000 if y < 0 else 0
    y = abs(y)
    e = max(int(mpmath.floor(mpmath.log(y, 2))), -126)
    q = int(mpmath.nint(y * mpmath.mpf(2) ** (23 - e)))
    v = q * mpmath.mpf(2) ** (e - 23)
    if v >= mpmath.mpf(2) ** 128:
        return s | 0x7F800000
    return s | bits32(float(v))


def pow32(x, y):
    x, y = mpmath.mpf(f32(x)), mpmath.mpf(f32(y))
    if x < 0:
        n = int(y)
        r = mpmath.power(-x, y)
        return round32(-r if n % 2 else r)
    return round32(mpmath.power(x, y))


rnd = random.Random(32)
inputs = []


def finite(v):
    return v & 0x7F800000 != 0x7F800000 and v & 0x7FFFFFFF != 0


# random bits
while len(inputs) < 200:
    x, y = rnd.getrandbits(31), rnd.getrandbits(32)
    if finite(x) and finite(y):
        inputs.append((x, y))
# moderate values
for _ in range(300):
    inputs.append((bits32(rnd.uniform(0, 10)), bits32(rnd.uniform(-30, 30))))
# x near 1 and large y
for _ in range(200):
    inputs.append((bits32(1 + rnd.uniform(-1e-3, 1e-3)), bits32(rnd.uniform(-1e5, 1e5))))
for _ in range(100):
    inputs.append((bits32(1 + rnd.uniform(-1e-6, 1e-6)), bits32(rnd.uniform(-1e8, 1e8))))
# negative x and integer y
for _ in range(100):
    inputs.append((bits32(-rnd.uniform(0, 10)), bits32(float(rnd.randint(-40, 40) or 1))))
# small integer y
for _ in range(100):
    inputs.append((bits32(rnd.uniform(0, 100)), bits32(float(rnd.randint(-10, 10) or 2))))
# exact results
for x in range(2, 12):
    for y in range(2, 8):
        inputs.append((bits32(float(x)), bits32(float(y))))
        inputs.append((bits32(float(x)), bits32(float(-y))))
# exact results on the midpoint of two adjacent Float32 values, e.g. (1+2**-12)**2 = 1 + 2**-11 + 2**-24
for e in range(12, 20):
    inputs.append((bits32(1 + 2.0**-e), bits32(2.0)))
    inputs.append((bits32(1 - 2.0**-e), bits32(2.0)))
# exact results o**n on the midpoint, where o is odd and o**n has 25 significant bits
mids = [(o, n) for n in range(3, 25) for o in range(3, 1 << 12, 2) if (o**n).bit_length() == 25]
for o, n in rnd.sample(mids, 40):
    e = rnd.randint(-3, 3)
    inputs.append((bits32(float(o) * 2.0**e), bits32(float(n))))
# exact results on the midpoint in the subnormal range, e.g. (3*2**-50)**3 = 13.5 * 2**-149
inputs.append((bits32(3 * 2.0**-50), bits32(3.0)))
inputs.append((bits32(2.0**-50), bits32(3.0)))
inputs.append((bits32(2.0**-100), bits32(1.5)))
# exact results on the midpoint with non-integer y, e.g. (s**2)**1.5 = s**3.
# They are not handled specially, and may not be correctly rounded.
for s in (257, 301, 321):
    if (s**3).bit_length() == 25:
        inputs.append((bits32(float(s * s)), bits32(1.5)))
# subnormal results, overflow, and underflow
for _ in range(50):
    inputs.append((bits32(rnd.uniform(1e-10, 1e-5)), bits32(rnd.uniform(4, 16))))
for y in (127.99, 128.0, -149.0, -149.5, -150.0, -150.01):
    inputs.append((bits32(2.0), bits32(y)))

with open("testdata/pow32.txt", "w") as f:
    for x, y in inputs:
        f.write(f"{x:08x} {y:08x} {pow32(x, y):08x}\n")
