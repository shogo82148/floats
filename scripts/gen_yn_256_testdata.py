#!/usr/bin/env python3
# Generates testdata/yn_256.txt.
# Each line contains the order n in decimal, the bits of x, and the correctly rounded Yn(x) in hexadecimal.
#
# Usage: python3 scripts/gen_yn_256_testdata.py

import random

import mpmath

P, EB = 236, 19
B = (1 << (EB - 1)) - 1
mpmath.mp.prec = 1000
rnd = random.Random(2562)


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
    cases = []
    for n in (2, 3, 5, 10, 20, 50, 100, 300, 1000):
        # x is log-uniform in the ranges relative to n: x < n (the result grows), x ~ n (the turning point),
        # and x > n (the result oscillates).
        for lo, hi, count in ((1e-3 * n, n, 8), (0.5 * n, 2 * n, 8), (n, 100 * n, 8), (100 * n, 1e7, 4)):
            lo = max(lo, 0.01)
            for _ in range(count):
                x = mpmath.mpf(lo) * mpmath.mpf(hi / lo) ** mpmath.mpf(rnd.random())
                cases.append((n, x))
    # the boundary of the asymptotic expansion: x = n**2 / 2 and x = 128
    for n in (16, 20, 100, 300, 1000):
        cases.append((n, max(mpmath.mpf(n * n) / 2, mpmath.mpf(128)) * (1 + mpmath.mpf(rnd.random()) / 100)))

    # the zeros of Y_n, which are searched near the McMahon approximation (k + n/2 - 1/4) pi
    # in low precision. The arguments need to be only close to the zeros.
    for n in (2, 10, 50):
        for k in (1, 2, 5, 20):
            guess = (k + n / 2 - 0.25) * mpmath.pi
            with mpmath.workprec(100):
                z = mpmath.findroot(lambda t: mpmath.bessely(n, t), (guess - 0.5, guess + 0.5), solver="anderson", tol=mpmath.mpf(10) ** -20)
            cases.append((n, +z))

    lines = []
    for n, x in cases:
        xb = enc(x)
        lines.append(f"{n} {xb:064x} {enc(mpmath.bessely(n, dec(xb))):064x}")
    with open("testdata/yn_256.txt", "w") as f:
        f.write("\n".join(lines) + "\n")


main()
