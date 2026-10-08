#!/usr/bin/env python3
# Generates testdata/y0_256.txt.
# Each line contains the bits of x and the correctly rounded Y0(x) in hexadecimal.
#
# Usage: python3 scripts/gen_y0_256_testdata.py

import random

import mpmath

P, EB = 236, 19
B = (1 << (EB - 1)) - 1
mpmath.mp.prec = 1000
rnd = random.Random(2560)


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


def y0(x):
    """returns Y0(x). mpmath.bessely is too slow for large x, so the Hankel expansion is used for x >= 200,
    where its smallest term is less than 2**-570."""
    if x < 200:
        return mpmath.bessely(0, x)
    with mpmath.workprec(mpmath.mp.prec + int(mpmath.log(x, 2)) + 64):
        p = q = mpmath.mpf(0)
        a = mpmath.mpf(1)
        for k in range(0, 400):
            term = a / x**k
            if abs(term) < mpmath.mpf(2) ** -600:
                break
            # a_k(0) = (-1)**k (1*3*...*(2k-1))**2 / (k! 8**k)
            if k % 2 == 0:
                p += (-1) ** (k // 2) * term
            else:
                q += (-1) ** ((k - 1) // 2 + 1) * term
            a = a * (2 * k + 1) ** 2 / ((k + 1) * 8)
        chi = x - mpmath.pi / 4
        return mpmath.sqrt(2 / (mpmath.pi * x)) * (p * mpmath.sin(chi) + q * mpmath.cos(chi))


def main():
    xs = []
    # tiny arguments
    for e in (-70000, -4000, -1000, -100, -40):
        xs.append(mpmath.mpf(1 + rnd.random()) * mpmath.mpf(10) ** e)
    # the power series, the Taylor series and the asymptotic expansion
    for lo, hi, n in ((1e-6, 2, 100), (2, 8, 100), (8, 128, 150), (128, 1000, 100), (1000, 1e6, 50)):
        for _ in range(n):
            xs.append(mpmath.mpf(lo) * mpmath.mpf(hi / lo) ** mpmath.mpf(rnd.random()))
    # around the boundaries and the zeros
    for b in (2, 8, 128):
        for d in (-1, 1):
            xs.append(mpmath.mpf(b) * (1 + d * mpmath.mpf(2) ** -240))
    for k in range(1, 30):
        z = mpmath.findroot(lambda t: mpmath.bessely(0, t), (k - 0.75) * mpmath.pi, tol=mpmath.mpf(10) ** -280)
        xs.append(z * (1 + mpmath.mpf(rnd.random() - 0.5) * 1e-6))
    # huge arguments
    for e in (30, 300, 30000, 70000):
        for _ in range(3):
            xs.append(mpmath.mpf(1 + rnd.random()) * mpmath.mpf(10) ** e)

    lines = []
    for x in xs:
        xb = enc(x)
        xr = dec(xb)
        lines.append(f"{xb:064x} {enc(y0(xr)):064x}")
    with open("testdata/y0_256.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
