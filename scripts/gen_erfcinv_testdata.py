#!/usr/bin/env python3
# Generates testdata/erfcinv128.txt, the correctly rounded erfcinv(x) for various 0 < x < 2.
# Each line contains the bits of x and the result in hexadecimal.
#
# Usage: python3 scripts/gen_erfcinv_testdata.py

import functools
import random
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf


def erfcinv(x):
    """returns erfcinv(x) for 0 < x < 2."""
    if x > 1:
        return -erfcinv(2 - x)
    if x >= mpmath.mpf(1) / 2:
        return mpmath.erfinv(1 - x)
    # Newton's method for ln erfc(y) = ln x, whose initial approximation is calculated with the low precision.
    lx = mpmath.log(x)
    with mpmath.workprec(60):
        y = mpmath.sqrt(-lx)
        for _ in range(60):
            l = mpmath.log(mpmath.erfc(y)) if y < 20 else -y * y - mpmath.log(y * mpmath.sqrt(mpmath.pi)) + mpmath.log1p(-1 / (2 * y * y))
            y = y + (l - lx) / (2 / mpmath.sqrt(mpmath.pi) * mpmath.exp(-y * y - l))
    for _ in range(12):
        e = mpmath.erfc(y)
        y = y + (mpmath.log(e) - lx) / (2 / mpmath.sqrt(mpmath.pi) * mpmath.exp(-y * y) / e)
    return y


def compute(v, P, EB):
    with mpmath.workprec(4 * P + 300):
        y = erfcinv(to_mpf(v, P, EB))
        if y == 0:
            return 0
        return round_bits(y, P, EB)


def gen(name, P, EB, seed):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)
    prec = 4 * P + 300

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def random_value(emin, emax):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax))

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(x, P, EB)

    inputs = []
    # the range of the Newton's method in the log domain: x < 2**-113, down to the subnormal numbers
    inputs += [random_value(-B + 1, -P - 2) for _ in range(150)]
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(30)]  # subnormal numbers
    inputs += [random_value(-P - 2, -P - 2 + 3) for _ in range(30)]
    # 2**-113 <= x <= 1/2
    inputs += [random_value(-P - 1, -2) for _ in range(150)]
    inputs += [random_value(-6, -2) for _ in range(100)]
    inputs += [random_value(-2, -2) for _ in range(40)]
    # 1/2 < x < 1, and 1 < x < 2
    inputs += [random_value(-1, -1) for _ in range(100)]
    inputs += [random_value(0, 0) for _ in range(150)]
    # close to 1 and 2
    with mpmath.workprec(prec):
        for k in range(1, P + 1, 3):
            t = mpmath.mpf(2) ** -k * (1 + mpmath.mpf(rnd.random()))
            inputs.append(round_bits(1 - t, P, EB))
            inputs.append(round_bits(1 + t, P, EB))
            inputs.append(round_bits(2 - t, P, EB))
        # around the boundaries
        for t in (mpmath.mpf(2) ** -(P + 1), mpmath.mpf(1) / 2, mpmath.mpf(1)):
            u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(t, 2))) - P)
            for d in range(-3, 4):
                inputs.append(round_bits(t + d * u, P, EB))
    inputs += [1]  # the smallest subnormal number
    inputs = [v for v in inputs if 0 < to_mpf_exact(v, P, EB) < 2]

    with Pool() as p:
        results = p.map(functools.partial(compute, P=P, EB=EB), inputs, chunksize=4)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, y in zip(inputs, results):
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


def to_mpf_exact(v, P, EB):
    with mpmath.workprec(4 * P + 300):
        return to_mpf(v, P, EB)


if __name__ == "__main__":
    gen("erfcinv128", 112, 15, 1281)
