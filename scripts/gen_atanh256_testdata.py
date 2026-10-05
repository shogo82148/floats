#!/usr/bin/env python3
# Generates testdata/atanh256.txt.
# Each line contains the bits of x and the correctly rounded atanh(x) in hexadecimal.
#
# Usage: python3 scripts/gen_atanh256_testdata.py

import random
import mpmath


def gen(name, P, EB, seed):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def dec(v):
        s = v >> (P + EB)
        e = (v >> P) & ((1 << EB) - 1)
        if e == 0:
            x = mpmath.mpf(v & ((1 << P) - 1)) * mpmath.mpf(2) ** (1 - B - P)
        else:
            x = mpmath.mpf((v & ((1 << P) - 1)) | (1 << P)) * mpmath.mpf(2) ** (e - B - P)
        return -x if s else x

    def round_bits(x):
        # round x to the nearest representable value.
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
        return enc(m, e, s)

    def nearest(x):
        with mpmath.workprec(4 * P + 64):
            return round_bits(x)

    def random_value(emin, emax, s=None):
        if s is None:
            s = rnd.getrandbits(1)
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    def random_subnormal(s=None):
        if s is None:
            s = rnd.getrandbits(1)
        frac = rnd.getrandbits(P)
        if frac == 0:
            frac = 1
        return frac | (s << (P + EB))

    # Atanh returns x for |x| < 2**-T, and uses the series for |x| < 2**-9.
    T = 119

    inputs = []
    # tiny arguments: atanh(x) ~ x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(20)]
    inputs += [random_value(-P - 2, -T - 4) for _ in range(40)]
    inputs += [random_subnormal() for _ in range(40)]
    inputs += [1, (1 << P) - 1, 1 | (1 << (P + EB)), ((1 << P) - 1) | (1 << (P + EB))]
    # |x| < 2**-9 (the series) and 2**-9 <= |x| < 1 (the logarithm)
    inputs += [random_value(-T - 3, -11) for _ in range(250)]
    inputs += [random_value(-12, -8) for _ in range(100)]
    inputs += [random_value(-8, -1) for _ in range(250)]
    inputs += [random_value(-3, -1) for _ in range(100)]

    with mpmath.workprec(4 * P + 64):
        # around the threshold (2**-T), where Atanh switches to a ~ x
        for e in range(-T - 5, -T + 2):
            for s in range(2):
                inputs += [enc(1 << P, e, s), enc((1 << P) + 1, e, s), enc((2 << P) - 1, e, s)]
        # close to 1, where atanh(x) is large and 1-x is exact in the same width
        for _ in range(150):
            d = rnd.randint(1, 1 << rnd.randint(1, P))
            for s in range(2):
                inputs.append(nearest((1 - mpmath.mpf(d) * mpmath.mpf(2) ** (-P - 1)) * (-1 if s else 1)))
        for d in range(1, 6):
            for s in range(2):
                inputs.append(nearest((1 - mpmath.mpf(d) * mpmath.mpf(2) ** (-P - 1)) * (-1 if s else 1)))
        # around 2**-10, 2**-9 and 0.5, where the algorithm switches
        two = mpmath.mpf(2)
        for center in (two ** -10, two ** -9, mpmath.mpf(1) / 2):
            for d in range(-3, 4):
                for s in (1, -1):
                    inputs.append(nearest(s * (center + d * center * two ** (-P - 2))))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            with mpmath.workprec(4 * P + 64):
                x = dec(v)
            with mpmath.workprec(4 * P + 64):
                y = round_bits(mpmath.atanh(x))
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


gen("atanh256", 236, 19, 256)
