#!/usr/bin/env python3
# Generates testdata/j0128.txt and testdata/j0256.txt, the correctly rounded J0(x) for various x.
# Each line contains the bits of x and the result in hexadecimal.
#
# Usage: python3 scripts/gen_j0_testdata.py [128|256]

import functools
import random
import sys
from multiprocessing import Pool

import mpmath

from gen_gamma_testdata import round_bits, to_mpf

P, EB = 236, 19
B = (1 << (EB - 1)) - 1


def j0(x):
    """returns J0(x) for x > 0 with the working precision."""
    if x < 2**20:
        return mpmath.besselj(0, x)
    # Hankel's asymptotic expansion, whose terms decrease until they are smaller than 2**-prec for x >= 2**20.
    prec = mpmath.mp.prec
    w = 1 / (8 * x)
    t = mpmath.mpf(1)
    p, q = mpmath.mpf(1), mpmath.mpf(0)
    k = 0
    while True:
        k += 1
        t = t * (2 * k - 1) ** 2 * w / k
        if t < mpmath.mpf(2) ** -(prec + 10):
            break
        sgn = -1 if (k // 2) % 2 else 1
        if k % 2 == 0:
            p += sgn * t
        else:
            q += sgn * t  # q is -Q
    chi = x - mpmath.pi / 4
    return mpmath.sqrt(2 / (mpmath.pi * x)) * (p * mpmath.cos(chi) + q * mpmath.sin(chi))


def init(p, eb):
    global P, EB, B
    P, EB = p, eb
    B = (1 << (EB - 1)) - 1


def compute(v):
    # x must be converted with enough precision.
    with mpmath.workprec(4 * P + 400 + max(0, (v >> P) - B)):
        x = to_mpf(v, P, EB)
        if x == 0:
            return round_bits(mpmath.mpf(1), P, EB)
        y = j0(x)
        if y == 0:
            return 0
        return round_bits(y, P, EB)


def main(name):
    rnd = random.Random(2562 if P == 236 else 1282)
    width = (P + EB + 1) // 4
    prec = 4 * P + 400

    def enc(m, e):
        return ((e + B) << P) | (m - (1 << P))

    def random_value(emin, emax):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax))

    def from_real(x):
        with mpmath.workprec(prec):
            return round_bits(x, P, EB)

    inputs = []
    # J0(x) is rounded to 1 for tiny x, and 1 - x**2/4
    tiny = -(P // 2)  # J0(x) = 1 - x**2/4 is rounded to 1 for x < 2**tiny
    inputs += [random_value(-B + 1, tiny - 1) for _ in range(20)]
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(5)]  # subnormal numbers
    inputs += [random_value(tiny - 4, tiny + 6) for _ in range(40)]
    # the Taylor series: x < 16
    inputs += [random_value(tiny + 6, 3) for _ in range(250)]
    # the Miller algorithm: 16 <= x < 110
    inputs += [from_real(mpmath.mpf(16) + mpmath.mpf(rnd.random()) * 94) for _ in range(200)]
    # the Hankel expansion: x >= 110
    inputs += [random_value(6, 12) for _ in range(120)]
    inputs += [from_real(mpmath.mpf(110) + mpmath.mpf(rnd.random()) * 400) for _ in range(100)]
    inputs += [random_value(13, 200) for _ in range(60)]
    inputs += [random_value(201, B - 1) for _ in range(30)]
    # the boundaries
    with mpmath.workprec(prec):
        for t in (mpmath.mpf(16), mpmath.mpf(110), mpmath.mpf(2) ** tiny):
            u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(t, 2))) - P)
            for d in range(-3, 4):
                inputs.append(round_bits(t + d * u, P, EB))
        # the zeros of J0, where the value is small. J0(x) is correctly rounded within an ulp of the zeros below 256.
        for k in range(1, 82):
            z = mpmath.besseljzero(0, k)
            u = mpmath.mpf(2) ** (int(mpmath.floor(mpmath.log(z, 2))) - P)
            for d in (-2, -1, 0, 1, 2):
                inputs.append(round_bits(z + d * u, P, EB))
            # |J0(x)| ~ 2**-60, where the fixed point arithmetic is no longer used
            for d in (-1, 1):
                inputs.append(round_bits(z + d * mpmath.mpf(2) ** -58 * (1 + mpmath.mpf(rnd.random())), P, EB))
        # J0(x) is not correctly rounded if |J0(x)| < 2**-70 for x >= 256, so that the points are apart from the zeros.
        for k in (100, 500, 1000, 3000):
            z = mpmath.besseljzero(0, k)
            for d in (-3, -1, 1, 3):
                inputs.append(round_bits(z + d * mpmath.mpf(2) ** -50, P, EB))
    # the largest values
    inputs += [enc((1 << (P + 1)) - 1, B), 1, enc(1 << P, B - 1)]
    inputs = list(dict.fromkeys(inputs))

    with Pool(initializer=init, initargs=(P, EB)) as p:
        results = p.map(compute, inputs, chunksize=2)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, y in zip(inputs, results):
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


if __name__ == "__main__":
    if (sys.argv[1] if len(sys.argv) > 1 else "256") == "128":
        init(112, 15)
        main("j0128")
    else:
        main("j0256")
