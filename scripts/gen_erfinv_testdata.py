#!/usr/bin/env python3
# Generates testdata/erfinv128.txt and testdata/erfinv256.txt, the correctly rounded erfinv(x) for various x.
# Each line contains the bits of x and the result in hexadecimal.
#
# Usage: python3 scripts/gen_erfinv_testdata.py [128|256]

import functools
import random
import sys
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf


def compute(v, P, EB):
    with mpmath.workprec(3 * P + 256):
        x = to_mpf(v, P, EB)
        y = mpmath.erfinv(x)
        if y == 0:
            return v & (1 << (P + EB))  # ±0
        return round_bits(y, P, EB)


def gen(name, P, EB, seed):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)
    prec = 4 * P + 256

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def random_value(emin, emax):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), rnd.getrandbits(1))

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(x, P, EB)

    inputs = []
    # tiny arguments: erfinv(x) ~ sqrt(pi)/2 x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(40)]
    inputs += [random_value(-P - 2, -64 if P == 112 else -130) for _ in range(40)]
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(10)]  # subnormal numbers
    # the boundary of the series: 2**-60 for Float128, and 2**-125 for Float256
    b0 = -125 if P == 236 else -60
    inputs += [random_value(b0 - 2, b0 + 2) for _ in range(40)]
    # the ranges of the calculation: [2**b0, 1/2), [1/2, 1)
    inputs += [random_value(b0 + 2, -2) for _ in range(150)]
    inputs += [random_value(-2, -1) for _ in range(100)]
    inputs += [random_value(-1, -1) for _ in range(150)]
    # close to 1
    with mpmath.workprec(prec):
        for k in range(1, P + 1, 3):
            t = mpmath.mpf(2) ** -k * (1 + mpmath.mpf(rnd.random()))
            if t < mpmath.mpf(2) ** -(P + 1):
                continue
            inputs.append(round_bits(1 - t, P, EB) ^ (rnd.getrandbits(1) << (P + EB)))
        # around the boundaries
        for t in (mpmath.mpf(2) ** b0, mpmath.mpf(1) / 2, mpmath.mpf(1) - mpmath.mpf(2) ** -(P + 1)):
            u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(t, 2))) - P)
            for d in range(-3, 4):
                inputs.append(round_bits(t + d * u, P, EB))
    inputs += [1]  # the smallest subnormal number
    inputs = [v for v in inputs if abs(to_mpf(v, P, EB)) < 1]

    with Pool() as p:
        results = p.map(functools.partial(compute, P=P, EB=EB), inputs, chunksize=4)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, y in zip(inputs, results):
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


if __name__ == "__main__":
    if (sys.argv[1] if len(sys.argv) > 1 else "128") == "128":
        gen("erfinv128", 112, 15, 1280)
    else:
        gen("erfinv256", 236, 19, 2560)
