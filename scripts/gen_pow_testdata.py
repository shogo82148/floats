#!/usr/bin/env python3
# Generates testdata/pow128.txt and testdata/pow256.txt, the correctly rounded a**b for various a and b.
# Each line contains the bits of a, b, and the result in hexadecimal.
#
# Usage: python3 scripts/gen_pow_testdata.py [128|256]

import functools
import math
import random
import sys
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf

P, EB = 112, 15
B = (1 << (EB - 1)) - 1
PREC = 4 * P + 4096  # enough to calculate the correct rounding of the results with up to 1024 multiplications
SMALLEST = 0  # the exponent of the smallest subnormal number is -(B - 1 + P)


def init(p, eb):
    global P, EB, B
    P, EB = p, eb
    B = (1 << (EB - 1)) - 1


def compute(ab):
    a, b = ab
    with mpmath.workprec(PREC):
        x, y = to_mpf(a, P, EB), to_mpf(b, P, EB)
        r = mpmath.power(x, y)
        if isinstance(r, mpmath.mpc):
            raise ValueError("complex")
        if r == 0:
            return 0
        return round_bits(r, P, EB)


def main(name):
    rnd = random.Random(1286 if P == 112 else 2566)
    width = (P + EB + 1) // 4

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def rand_a(emin, emax, s=0):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    def from_real(x):
        with mpmath.workprec(PREC):
            return round_bits(mpmath.mpf(x), P, EB)

    def from_int(n):
        with mpmath.workprec(PREC):
            return round_bits(mpmath.mpf(n), P, EB)

    cases = []
    # small integer powers
    for _ in range(150):
        cases.append((rand_a(-40, 40, rnd.getrandbits(1)), from_int(rnd.randint(-20, 20))))
    for _ in range(100):
        cases.append((rand_a(-12, 12), from_int(rnd.randint(-1024, 1024))))
    for _ in range(40):
        cases.append((rand_a(-(B - 380), B - 380, rnd.getrandbits(1)), from_int(rnd.randint(2, 6))))
    # a**n is exactly the midpoint of two adjacent Float128 values
    for n in (2, 3, 4, 5, 6, 7):
        lo = int(2 ** ((P + 1.0) / n))
        hi = int(2 ** ((P + 2.0) / n))
        for _ in range(8):
            m = rnd.randint(lo, hi) | 1
            cases.append((from_int(m), from_int(n)))
            cases.append((from_int(m) | (1 << (P + EB)), from_int(n)))  # negative
    # a**n is exactly the midpoint of two subnormal numbers: m**n 2**-(B - 1 + P + 1), where n is the smallest odd divisor
    tie = B - 1 + P + 1
    n = next(d for d in range(3, tie + 1, 2) if tie % d == 0)
    for m in (3, 5, 7, 9, 11, 13, 15, 17, 63):
        cases.append((from_real(mpmath.mpf(m) * mpmath.mpf(2) ** -(tie // n)), from_int(n)))
    # powers of two, whose exponents are exact
    for e in (-3, -1, 1, 2, 5, 100, -(B - 380), B - 380):
        for n in (2, 3, 63, 1000, 5000, 100000, 1 << 40):
            cases.append((from_real(mpmath.mpf(2) ** e), from_int(n)))
            cases.append((from_real(mpmath.mpf(2) ** e), from_int(-n)))
    # the exact products have a little more than 2P+2 bits: they are not the midpoints, but the rounding errors of
    # the successive squarings are not negligible.
    for n in (2, 3, 4, 5, 6):
        for bits in range((2 * P + 2) // n - 1, (2 * P + 2) // n + 3):
            if bits > P + 1:
                continue
            for _ in range(6):
                cases.append((from_int(rnd.getrandbits(bits) | (1 << (bits - 1)) | 1), from_int(n)))
    # a close to 1, and large powers
    for k in range(10, P, 7):
        for _ in range(3):
            with mpmath.workprec(PREC):
                a = round_bits(1 + mpmath.mpf(rnd.random() + 0.5) * mpmath.mpf(2) ** -k, P, EB)
                a2 = round_bits(1 - mpmath.mpf(rnd.random() * 0.5 + 0.1) * mpmath.mpf(2) ** -k, P, EB)
            cases.append((a, from_int(rnd.randint(1, 1 << min(k + 8, 60)))))
            cases.append((a2, from_int(rnd.randint(1, 1 << min(k + 8, 60)))))
            cases.append((a, from_int(-rnd.randint(1, 1 << min(k + 8, 60)))))
    # the nearest values to 1 with the huge exponents: the results are about e, 1/e, and e**1.5
    one = from_int(1)
    for a, y in ((one + 1, 1 << P), (one + 1, 1 << (P + 1)), (one + 1, 3 << (P - 1)), (one - 1, 1 << P), (one - 1, 1 << (P + 1)), (one + 1, -(1 << P))):
        cases.append((a, from_int(y)))
    # huge exponents, which are integers, and the results are not trivial
    for k in range(64, P, 6):
        with mpmath.workprec(PREC):
            a = round_bits(1 + mpmath.mpf(rnd.random() + 0.5) * mpmath.mpf(2) ** -(k - 5), P, EB)
        cases.append((a, from_int(rnd.getrandbits(k - 3) | (1 << (k - 4)))))
    # non-integer exponents
    for _ in range(250):
        cases.append((rand_a(-30, 30), rand_a(-6, 5, rnd.getrandbits(1))))
    for _ in range(60):
        cases.append((rand_a(-1000, 1000), rand_a(-14, -4, rnd.getrandbits(1))))
    for _ in range(40):
        cases.append((rand_a(-1, 1), rand_a(-4, 12, rnd.getrandbits(1))))
    for _ in range(40):
        cases.append((rand_a(-(B - 80), B - 80, rnd.getrandbits(1)), rand_a(-110, -80, rnd.getrandbits(1)) if P == 112 else rand_a(-125, -90, rnd.getrandbits(1))))
    # perfect powers and the small fractions
    for a, y in ((4, 1.5), (8, 1 / 3), (9, 0.5), (27, 2 / 3), (16, 0.25), (16, -0.25), (4, -0.5), (2, -0.5), (10, 0.1)):
        cases.append((from_real(a), from_real(y)))
    # overflow and underflow
    for base in (2, 3, 10, 0.5, 0.1):
        with mpmath.workprec(PREC):
            for target in (B + 1, -(B - 1 + P), -(B - 1 + P) - 1, -(B - 1), B - 0.01):
                y = mpmath.mpf(target) / mpmath.log(base, 2)
                bb = round_bits(y, P, EB)
                for d in range(-2, 3):
                    cases.append((from_real(base), bb + d))
    # subnormal numbers and the extreme values
    for _ in range(20):
        cases.append((rnd.getrandbits(rnd.randint(1, P)) | 1, rand_a(-3, 3)))
    cases.append((1, from_int(2)))
    cases.append((from_real(2), from_int(B)))
    cases.append((from_real(2), from_int(B + 1)))
    cases.append((from_real(2), from_int(-(B - 1 + P))))
    cases.append((from_real(2), from_int(-(B - 1 + P) - 1)))
    cases.append((from_real(2) | (1 << (P + EB)), from_int(3)))
    cases.append((from_real(-3), from_int(1023)))
    cases.append((from_real(-3), from_int(-1024)))

    # exclude the special cases, which are tested elsewhere, and the NaNs
    def ok(case):
        a, b = case
        x, y = to_mpf(a, P, EB), to_mpf(b, P, EB)
        if x == 0 or y == 0 or abs(x) == 1:
            return False
        if x < 0 and y != int(y):
            return False
        return True

    with mpmath.workprec(PREC):
        cases = [c for c in cases if ok(c)]
    with Pool(initializer=init, initargs=(P, EB)) as p:
        results = p.map(compute, cases, chunksize=4)
    with open(f"testdata/{name}.txt", "w") as f:
        for (a, b), r in zip(cases, results):
            f.write(f"{a:0{width}x} {b:0{width}x} {r:0{width}x}\n")


if __name__ == "__main__":
    if (sys.argv[1] if len(sys.argv) > 1 else "128") == "128":
        main("pow128")
    else:
        P, EB = 236, 19
        B = (1 << (EB - 1)) - 1
        main("pow256")
