#!/usr/bin/env python3
# Generates testdata/lgamma128.txt.
# Each line contains the bits of x, the correctly rounded Lgamma(x), and the sign of Gamma(x), 1 or -1 in decimal.
#
# Usage: python3 scripts/gen_lgamma_testdata.py

import functools
import random
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf


def compute(v, P, EB):
    """returns the bits and the sign of Lgamma(x) for the bits v of x, or None if x is a pole."""
    B = (1 << (EB - 1)) - 1
    with mpmath.workprec(4 * P + 256 + B.bit_length() * 0 + 2 * EB + 600):
        x = to_mpf(v, P, EB)
        if x == 0 or (x < 0 and x == mpmath.floor(x)):
            return None
        if x > 0:
            y = mpmath.loggamma(x)
            sign = 1
        else:
            # |Gamma(x)| = pi / (|sin(pi x)| Gamma(1-x))
            y = mpmath.log(mpmath.pi) - mpmath.log(abs(mpmath.sin(mpmath.pi * x))) - mpmath.loggamma(1 - x)
            sign = -1 if int(mpmath.floor(x)) % 2 != 0 else 1
        if y == 0:
            return 0, sign
        return round_bits(y, P, EB), sign


def gen(name, P, EB, seed):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)
    prec = 4 * P + 256

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def random_value(emin, emax, s=0):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(x, P, EB)

    inputs = []
    # tiny arguments: Lgamma(x) ~ -log(x)
    inputs += [random_value(-B + 1, -P - 3, rnd.getrandbits(1)) for _ in range(60)]
    inputs += [random_value(-P - 2, -20, rnd.getrandbits(1)) for _ in range(40)]
    inputs += [random_value(-60, -35, rnd.getrandbits(1)) for _ in range(60)]  # around 2**-40, where the series is switched
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(12)]  # subnormal numbers
    # positive arguments
    inputs += [random_value(-19, -1) for _ in range(100)]  # (0, 1)
    inputs += [random_value(0, 1) for _ in range(150)]  # [1, 4)
    inputs += [random_value(2, 5) for _ in range(150)]  # [4, 64)
    inputs += [random_value(5, 10) for _ in range(150)]  # [32, 2048)
    inputs += [random_value(11, 14) for _ in range(100)]  # [2048, 2**15)
    inputs += [random_value(15, 60) for _ in range(100)]
    inputs += [random_value(61, 14000) for _ in range(60)]
    # negative arguments
    inputs += [random_value(-30, -1, 1) for _ in range(80)]
    inputs += [random_value(0, 4, 1) for _ in range(150)]
    inputs += [random_value(4, 6, 1) for _ in range(100)]
    inputs += [random_value(6, 10, 1) for _ in range(100)]
    inputs += [random_value(11, 14, 1) for _ in range(100)]
    inputs += [random_value(15, 60, 1) for _ in range(60)]
    inputs += [random_value(61, 110, 1) for _ in range(40)]
    # the integers and the half-integers
    for k in [1, 2, 3, 4, 5, 10, 22, 23, 24, 25, 30, 47, 48, 49, 100, 1000, 2047, 2048, 2049, 32767, 32768, 32769, 100000]:
        inputs.append(from_real(mpmath.mpf(k)))
        inputs.append(from_real(mpmath.mpf(k) + mpmath.mpf(1) / 2))
        inputs.append(from_real(-mpmath.mpf(k) - mpmath.mpf(1) / 2))
    with mpmath.workprec(prec + 400):
        # near the zeros of Lgamma at 1 and 2, where the relative accuracy matters
        for c in (mpmath.mpf(1), mpmath.mpf(2)):
            for e in (-1, -3, -5, -6, -7, -10, -20, -50, -100, -110, -112):
                for sgn in (1, -1):
                    inputs.append(round_bits(c + sgn * mpmath.mpf(2) ** e * (1 + mpmath.mpf(rnd.random())), P, EB))
            for d in (1, 2, 3):
                inputs.append(round_bits(c + d * mpmath.mpf(2) ** (-P), P, EB))
                inputs.append(round_bits(c - d * mpmath.mpf(2) ** (-P - 1), P, EB))
        # near the negative zeros of Lgamma
        for n in range(2, 40):
            for guess in (mpmath.mpf(-n) - 1 / mpmath.factorial(n), mpmath.mpf(-n) + 1 / mpmath.factorial(n)):
                try:
                    r = mpmath.findroot(lambda t: mpmath.log(abs(mpmath.gamma(t))), guess, tol=mpmath.mpf(2) ** -(prec + 300), maxsteps=200)
                except Exception:
                    continue
                if abs(r - guess) > mpmath.mpf(1) / 4:
                    continue
                base = round_bits(r, P, EB)
                inputs += [base + d for d in (-2, -1, 0, 1, 2)]
        # near the negative integers
        for n in [1, 2, 3, 10, 50, 171, 500, 1000, 2000, 2047, 2048, 5000, 32767, 32768, 100000]:
            for e in (-1, -5, -20, -60, -100):
                for sgn in (1, -1):
                    x = -mpmath.mpf(n) + sgn * mpmath.mpf(2) ** e * (1 + mpmath.mpf(rnd.random()))
                    inputs.append(round_bits(x, P, EB))
        # around the segments of the calculation
        for t in (24, 48, 2048, 32768):
            for d in (-3, -2, -1, 0, 1, 2, 3):
                inputs.append(round_bits(mpmath.mpf(t) + d * mpmath.mpf(2) ** (-P + max(0, int(mpmath.log(t, 2)))), P, EB))
                inputs.append(round_bits(-mpmath.mpf(t) - mpmath.mpf(1) / 3 + d * mpmath.mpf(2) ** (-P + max(0, int(mpmath.log(t, 2)))), P, EB))
        # near the overflow threshold of Lgamma
        top = (mpmath.mpf(2) ** (B + 1)) * (1 - mpmath.mpf(2) ** (-P - 2))
        guess = top / 11000
        for _ in range(60):
            guess = top / (mpmath.log(guess) - 1)
        lo, hi = guess / 2, guess * 2
        for _ in range(prec + 100):
            mid = (lo + hi) / 2
            if mpmath.loggamma(mid) < top:
                lo = mid
            else:
                hi = mid
        thr = lo
        base = round_bits(thr, P, EB)
        inputs += [base + d for d in range(-3, 4)]
    # the largest finite value
    inputs += [enc((1 << (P + 1)) - 1, B), 1, 1 | (1 << (P + EB))]

    with Pool() as p:
        results = p.map(functools.partial(compute, P=P, EB=EB), inputs, chunksize=8)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, r in zip(inputs, results):
            if r is not None:
                y, sign = r
                f.write(f"{v:0{width}x} {y:0{width}x} {sign}\n")


if __name__ == "__main__":
    gen("lgamma128", 112, 15, 128)
