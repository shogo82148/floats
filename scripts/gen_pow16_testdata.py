#!/usr/bin/env python3
# Generates testdata/pow16.txt.
# Each line contains the bits of x, y, and the correctly rounded x**y in hexadecimal.
#
# Usage: python3 scripts/gen_pow16_testdata.py

import random
import struct
from fractions import Fraction

import mpmath

mpmath.mp.prec = 600


def f16(v):
    return struct.unpack("<e", struct.pack("<H", v))[0]


def bits16(f):
    return struct.unpack("<H", struct.pack("<e", f))[0]


def round16(v):
    # round the exact value v (Fraction or mpf) to the nearest Float16 value, ties to even.
    if v == 0:
        return 0
    s = 0x8000 if v < 0 else 0
    v = abs(mpmath.mpf(v.numerator) / v.denominator if isinstance(v, Fraction) else v)
    e = max(int(mpmath.floor(mpmath.log(v, 2))), -14)
    q = int(mpmath.nint(v * mpmath.mpf(2) ** (10 - e)))
    r = q * mpmath.mpf(2) ** (e - 10)
    if r > 65504:
        return s | 0x7C00
    return s | bits16(float(r))


def pow16(x, y):
    fx, fy = Fraction(f16(x)), Fraction(f16(y))
    if fx < 0:
        r = mpmath.power(-f16(x), f16(y))
        return round16(-r if int(fy) % 2 else r)
    return round16(mpmath.power(f16(x), f16(y)))


def is_midpoint(v):
    # reports whether the exact value v (Fraction) is the midpoint of two adjacent Float16 values.
    v = abs(v)
    e = max(v.numerator.bit_length() - v.denominator.bit_length(), -14)
    while Fraction(2) ** e > v:
        e -= 1
    while Fraction(2) ** (e + 1) <= v:
        e += 1
    e = max(e, -14)
    u = v / Fraction(2) ** (e - 10)
    return u.denominator == 2


inputs = []

# exact results x**y on the midpoint, where y = n/2**q
for xb in range(1, 0x7C00):
    x = Fraction(f16(xb))
    for q in range(5):
        # root = x**(1/2**q) if it is exact
        num, den = x.numerator, x.denominator
        rn, rd = round(num ** (1 / 2**q)), round(den ** (1 / 2**q))
        if rn ** (2**q) != num or rd ** (2**q) != den:
            break
        root = Fraction(rn, rd)
        for n in range(-64, 65):
            if n == 0 or n % 2 == 0 and q > 0:
                continue
            y = Fraction(n, 2**q)
            yf = float(y)
            if bits16(yf) is None or Fraction(f16(bits16(yf))) != y:
                continue
            v = root**n
            if v > 65520 or v < Fraction(1, 2**26):
                continue
            if is_midpoint(v):
                inputs.append((xb, bits16(yf)))
                inputs.append((xb | 0x8000, bits16(yf))) if q == 0 else None
print("exact midpoints:", len(inputs))

# random pairs
rnd = random.Random(16)
for _ in range(2000):
    inputs.append((rnd.getrandbits(16), rnd.getrandbits(16)))

with open("testdata/pow16.txt", "w") as f:
    for x, y in inputs:
        fx, fy = f16(x), f16(y)
        if fx != fx or fy != fy or abs(fx) == float("inf") or abs(fy) == float("inf") or fx == 0 or fy == 0:
            continue  # special cases are tested separately
        if fx < 0 and fy != int(fy):
            continue
        f.write(f"{x:04x} {y:04x} {pow16(x, y):04x}\n")
