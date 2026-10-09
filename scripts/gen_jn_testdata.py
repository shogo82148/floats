#!/usr/bin/env python3
# Generates testdata/jn256.txt, the correctly rounded Jn(x) for various n and x.
# Each line contains n in decimal, and the bits of x and the result in hexadecimal.
#
# Usage: python3 scripts/gen_jn_testdata.py

import functools
import math
import random
from multiprocessing import Pool

import mpmath

import gen_j0_testdata as g
from gen_gamma_testdata import round_bits, to_mpf

P, EB = 236, 19
B = (1 << (EB - 1)) - 1


def compute(nx):
    n, v = nx
    g.init(P, EB, abs(n))
    with mpmath.workprec(4 * P + 400 + max(0, ((v >> P) & ((1 << EB) - 1)) - B)):
        x = to_mpf(v, P, EB)
        if x == 0:
            return 0
        k = abs(n)
        ax = abs(x)
        # Jn(-x) = (-1)**n Jn(x), J(-n)(x) = (-1)**n Jn(x)
        if ax >= 2**20 and ax >= 16 * k * k:
            y = g.bessel(ax)  # Hankel's expansion
        elif ax < 16:
            y = g.bessel(ax)  # the Taylor series
        else:
            y = mpmath.besselj(k, ax)
        if k % 2 == 1 and (x < 0) != (n < 0):
            y = -y
        if y == 0:
            return 0
        return round_bits(y, P, EB)


def main():
    rnd = random.Random(2563)
    width = (P + EB + 1) // 4
    prec = 4 * P + 400

    def enc(m, e):
        return ((e + B) << P) | (m - (1 << P))

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(mpmath.mpf(x), P, EB)

    def uniform(lo, hi):
        with mpmath.workprec(prec):
            return from_real(mpmath.mpf(lo) + mpmath.mpf(rnd.random()) * (mpmath.mpf(hi) - mpmath.mpf(lo)))

    cases = []
    for n in (2, 3, 4, 5, 7, 10, 20, 50, 100, 200, 500, 1000, 2000):
        tr = math.sqrt(2 * (n + 1))  # the boundary of the Taylor series
        xs = []
        xs += [enc(rnd.getrandbits(P) | (1 << P), rnd.randint(-200, -2)) for _ in range(3)]
        xs += [uniform(0.01, tr) for _ in range(5)]
        xs += [from_real(tr * f) for f in (0.999999, 1.0, 1.000001)]
        xs += [uniform(tr, max(n, tr + 1)) for _ in range(5)]
        xs += [uniform(n, 2 * n) for _ in range(5)]
        xs += [uniform(2 * n, 8 * n + 10) for _ in range(4)]
        # around the boundary of Hankel's expansion: x >= max(110, 2 n**2)
        for f in (0.9, 0.999, 1.0, 1.001, 3, 100):
            xs.append(from_real(max(110, 2 * n * n) * f))
        xs += [from_real(2.0**60 + rnd.random() * 1000), from_real(2.0**500 + 2.0**440 * rnd.random())]
        for v in xs:
            sign_x = rnd.getrandbits(1)
            sign_n = rnd.getrandbits(1)
            cases.append((-n if sign_n else n, v | (sign_x << (P + EB))))
    cases = list(dict.fromkeys(cases))
    with Pool() as p:
        results = p.map(compute, cases, chunksize=2)
    with open("testdata/jn256.txt", "w") as f:
        for (n, v), y in zip(cases, results):
            f.write(f"{n} {v:0{width}x} {y:0{width}x}\n")


if __name__ == "__main__":
    main()
