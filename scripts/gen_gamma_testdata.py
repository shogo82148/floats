#!/usr/bin/env python3
# Generates testdata/gamma128.txt.
# Each line contains the bits of x and the correctly rounded Gamma(x) in hexadecimal.
#
# Usage: python3 scripts/gen_gamma_testdata.py

import functools
import random
from multiprocessing import Pool

import mpmath


def round_bits(x, P, EB):
    """rounds x to the nearest representable value and returns its bits."""
    B = (1 << (EB - 1)) - 1
    s = 1 if x < 0 else 0
    x = abs(x)
    if x == 0:
        return s << (P + EB)
    e = max(int(mpmath.floor(mpmath.log(x, 2))), 1 - B)
    m = int(mpmath.nint(x * mpmath.mpf(2) ** (P - e)))
    if m >> (P + 1):
        m >>= 1
        e += 1
    if e > B:
        return (s << (P + EB)) | (((1 << EB) - 1) << P)  # infinity
    if m >> P == 0:
        return m | (s << (P + EB))  # subnormal
    return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))


def to_mpf(v, P, EB):
    B = (1 << (EB - 1)) - 1
    s = v >> (P + EB)
    e = (v >> P) & ((1 << EB) - 1)
    f = v & ((1 << P) - 1)
    x = mpmath.mpf(f) * mpmath.mpf(2) ** (1 - B - P) if e == 0 else mpmath.mpf(f | (1 << P)) * mpmath.mpf(2) ** (e - B - P)
    return -x if s else x


def compute(v, P, EB):
    """returns the bits of Gamma(x) for the bits v of x, or None if Gamma(x) is NaN."""
    with mpmath.workprec(4 * P + 256):
        x = to_mpf(v, P, EB)
        if x == 0 or (x < 0 and x == mpmath.floor(x)):
            return None
        if x > 2000:
            return ((1 << EB) - 1) << P  # +Inf
        if x < -2000:
            return None  # the result underflows to zero; its sign is not tested
        return round_bits(mpmath.gamma(x), P, EB)


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
    # tiny arguments: Gamma(x) ~ 1/x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(30)]
    inputs += [random_value(-P - 2, -20) for _ in range(30)]
    inputs += [random_value(-B + 1, -B + 3) for _ in range(10)]  # close to the overflow of 1/x
    inputs += [random_value(-60, -41, rnd.getrandbits(1)) for _ in range(40)]  # around 2**-40, where the series is switched
    inputs += [random_value(-41, -38, rnd.getrandbits(1)) for _ in range(40)]
    inputs += [rnd.getrandbits(rnd.randint(1, P)) | 1 for _ in range(12)]  # subnormal numbers
    # positive arguments
    inputs += [random_value(-19, -1) for _ in range(120)]  # (0, 1)
    inputs += [random_value(0, 1) for _ in range(150)]  # [1, 4)
    inputs += [random_value(2, 5) for _ in range(150)]  # [4, 64)
    inputs += [random_value(5, 10) for _ in range(150)]  # [32, 2048)
    # negative arguments
    inputs += [random_value(-30, -1, 1) for _ in range(80)]
    inputs += [random_value(0, 4, 1) for _ in range(150)]
    inputs += [random_value(4, 6, 1) for _ in range(100)]
    inputs += [random_value(6, 10, 1) for _ in range(180)]
    # the integers and the half-integers
    for k in [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 20, 21, 22, 30, 50, 54, 55, 56, 100, 171, 172, 1000, 1754, 1755]:
        inputs.append(from_real(mpmath.mpf(k)))
        inputs.append(from_real(mpmath.mpf(k) + mpmath.mpf(1) / 2))
        inputs.append(from_real(-mpmath.mpf(k) - mpmath.mpf(1) / 2))
    with mpmath.workprec(prec):
        # near the overflow threshold of Gamma
        thr = mpmath.findroot(lambda t: mpmath.loggamma(t) - (B + 1) * mpmath.log(2), mpmath.mpf(1755))
        base = round_bits(thr, P, EB)
        inputs += [base + d for d in range(-3, 4)]
        # near the negative integers
        for n in [1, 2, 3, 10, 50, 171, 500, 1000, 1700, 1780, 1790]:
            for e in (-1, -5, -20, -60, -100):
                for sgn in (1, -1):
                    x = -mpmath.mpf(n) + sgn * mpmath.mpf(2) ** e * (1 + mpmath.mpf(rnd.random()))
                    inputs.append(round_bits(x, P, EB))
        # close to 1, 2 and the minimum of Gamma
        for c in (mpmath.mpf(1), mpmath.mpf(2), mpmath.mpf("1.4616321449683623412626595423257213284681962")):
            for e in (-2, -10, -50, -100, -112):
                for sgn in (1, -1):
                    inputs.append(round_bits(c + sgn * mpmath.mpf(2) ** e, P, EB))
    # the largest finite value and the smallest subnormal numbers
    inputs += [enc((1 << (P + 1)) - 1, B), 1, 1 | (1 << (P + EB))]

    with Pool() as p:
        results = p.map(functools.partial(compute, P=P, EB=EB), inputs, chunksize=8)
    with open(f"testdata/{name}.txt", "w") as f:
        for v, y in zip(inputs, results):
            if y is not None:
                f.write(f"{v:0{width}x} {y:0{width}x}\n")


if __name__ == "__main__":
    gen("gamma128", 112, 15, 128)
