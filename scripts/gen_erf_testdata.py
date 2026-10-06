#!/usr/bin/env python3
# Generates testdata/erf128.txt, testdata/erf256.txt, testdata/erfc128.txt, and testdata/erfc256.txt.
# Each line contains the bits of x and the correctly rounded erf(x) or erfc(x) in hexadecimal.
#
# Usage: python3 scripts/gen_erf_testdata.py [128|256] [erf|erfc]

import functools
import random
import sys
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf


def compute(v, P, EB, fn):
    with mpmath.workprec(4 * P + 256):
        x = to_mpf(v, P, EB)
        if fn == "erfc" and abs(x) > 1 << 12:
            # mpmath cannot calculate erfc(x) for huge x; it is rounded to 0 or 2.
            return round_bits(mpmath.mpf(2) if x < 0 else mpmath.mpf(0), P, EB)
        y = getattr(mpmath, fn)(x)
        if y == 0:
            return v & (1 << (P + EB)) if fn == "erf" else 0  # ±0
        return round_bits(y, P, EB)


def gen(name, P, EB, seed, fn="erf"):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)
    prec = 4 * P + 256
    big = P == 236

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def random_value(emin, emax, s=None):
        if s is None:
            s = rnd.getrandbits(1)
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(x, P, EB)

    inputs = []
    # tiny arguments: erf(x) ~ 2 x / sqrt(pi)
    inputs += [random_value(-B + 1, -P - 3) for _ in range(60)]
    inputs += [random_value(-P - 2, -70) for _ in range(40)]
    inputs += [random_value(-70, -9) for _ in range(120)]
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(12)]  # subnormal numbers
    # the ranges of the calculation: [2**-8, 1), [1, 2), [2, 4), [4, 7), [7, 9)
    inputs += [random_value(-8, -1) for _ in range(150)]
    inputs += [random_value(0, 0) for _ in range(120)]
    inputs += [random_value(1, 1) for _ in range(120)]
    inputs += [random_value(2, 2) for _ in range(100)]
    inputs += [from_real(mpmath.mpf(4) + mpmath.mpf(rnd.random()) * 3) for _ in range(100)]
    inputs += [from_real(mpmath.mpf(7) + mpmath.mpf(rnd.random()) * (6 if big else 1.8)) for _ in range(100 if not big else 200)]
    inputs += [random_value(4, 12) for _ in range(20)]
    with mpmath.workprec(prec):
        # around the boundaries of the ranges
        for t in (mpmath.mpf(2) ** -8, mpmath.mpf(1), mpmath.mpf(2), mpmath.mpf(4), mpmath.mpf(6), mpmath.mpf(7), mpmath.mpf(8)) + ((mpmath.mpf(12),) if big else ()):
            u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(t, 2))) - P)
            for d in range(-3, 4):
                inputs.append(round_bits(t + d * u, P, EB))
        # near the threshold where erf(x) is rounded to 1
        lo, hi = mpmath.mpf(5), mpmath.mpf(12)
        for _ in range(prec + 100):
            m = (lo + hi) / 2
            if mpmath.erfc(m) > mpmath.mpf(2) ** (-P - 2):
                lo = m
            else:
                hi = m
        base = round_bits(hi, P, EB)
        inputs += [base + d for d in range(-3, 4)]
    if fn == "erfc":
        # large arguments: [4, 8), [8, 16), [16, 32), [32, 64), and close to the underflow
        inputs += [from_real(mpmath.mpf(4) + mpmath.mpf(rnd.random()) * 4) for _ in range(100)]
        inputs += [from_real(mpmath.mpf(8) + mpmath.mpf(rnd.random()) * 8) for _ in range(100)]
        inputs += [from_real(mpmath.mpf(16) + mpmath.mpf(rnd.random()) * 16) for _ in range(100)]
        inputs += [from_real(mpmath.mpf(32) + mpmath.mpf(rnd.random()) * 32) for _ in range(60)]
        if not big:
            inputs += [from_real(mpmath.mpf(64) + mpmath.mpf(rnd.random()) * 36) for _ in range(40)]
            inputs += [from_real(mpmath.mpf(104) + mpmath.mpf(rnd.random()) * 3) for _ in range(60)]
        else:
            inputs += [from_real(mpmath.mpf(64) + mpmath.mpf(rnd.random()) * 359) for _ in range(60)]
            inputs += [from_real(mpmath.mpf(423) + mpmath.mpf(rnd.random()) * 3.4) for _ in range(60)]
            inputs += [random_value(9, 20) for _ in range(10)]  # erfc(x) is rounded to 0
        with mpmath.workprec(prec):
            for t in (mpmath.mpf(3), mpmath.mpf(4), mpmath.mpf(8), mpmath.mpf(16), mpmath.mpf(32)) + ((mpmath.mpf(256),) if big else ()):
                u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(t, 2))) - P)
                for d in range(-3, 4):
                    inputs.append(round_bits(t + d * u, P, EB))
            # near the threshold where erfc(x) is rounded to 0
            if not big:
                lo, hi, tiny = mpmath.mpf(100), mpmath.mpf(110), mpmath.mpf(2) ** (-(1 << 14) - P + 1 - 2)
            else:
                # the half of the smallest subnormal number
                lo, hi, tiny = mpmath.mpf(420), mpmath.mpf(430), mpmath.mpf(2) ** (2 - (1 << (EB - 1)) - P - 1)
            for _ in range(prec + 100):
                m = (lo + hi) / 2
                if mpmath.erfc(m) > tiny:
                    lo = m
                else:
                    hi = m
            base = round_bits(hi, P, EB)
            inputs += [base + d for d in range(-3, 4)]
            # near the threshold where erfc(-x) is rounded to 2
            lo, hi = mpmath.mpf(5), mpmath.mpf(14 if big else 12)
            for _ in range(prec + 100):
                m = (lo + hi) / 2
                if mpmath.erfc(m) > mpmath.mpf(2) ** (-P - 1):
                    lo = m
                else:
                    hi = m
            base = round_bits(hi, P, EB)
            inputs += [base + d for d in range(-3, 4)]
    # the largest finite value
    inputs += [enc((1 << (P + 1)) - 1, B), 1, 0]
    inputs = [v ^ (rnd.getrandbits(1) << (P + EB)) if i % 2 else v for i, v in enumerate(inputs)]  # half of them are negative

    with Pool() as p:
        results = p.map(functools.partial(compute, P=P, EB=EB, fn=fn), inputs, chunksize=8)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, y in zip(inputs, results):
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


if __name__ == "__main__":
    fn = sys.argv[2] if len(sys.argv) > 2 else "erf"
    if (sys.argv[1] if len(sys.argv) > 1 else "128") == "128":
        gen(fn + "128", 112, 15, 128, fn)
    else:
        gen(fn + "256", 236, 19, 256, fn)
