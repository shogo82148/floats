#!/usr/bin/env python3
# Generates testdata/asinh256.txt.
# Each line contains the bits of x and the correctly rounded asinh(x) in hexadecimal.
#
# Usage: python3 scripts/gen_asinh256_testdata.py

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

    # Asinh returns x for |x| < 2**-T, and approximates log(2|x|) for |x| >= 2**H.
    T, H = 118, 128

    inputs = []
    # tiny arguments: asinh(x) ~ x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(20)]
    inputs += [random_value(-P - 2, -T - 4) for _ in range(40)]
    inputs += [random_subnormal() for _ in range(40)]
    inputs += [1, (1 << P) - 1, 1 | (1 << (P + EB)), ((1 << P) - 1) | (1 << (P + EB))]
    # small-to-medium arguments across the branches of Asinh:
    # |a| < 2**-T, t = a + a**2/(1+sqrt(1+a**2)) < 2**-8, |a| < 1, |a| < 2**H, and |a| >= 2**H
    inputs += [random_value(-T - 3, -11) for _ in range(250)]
    inputs += [random_value(-12, -7) for _ in range(100)]
    inputs += [random_value(-6, -1) for _ in range(100)]
    inputs += [random_value(0, 1) for _ in range(100)]
    inputs += [random_value(2, 40) for _ in range(100)]
    inputs += [random_value(41, H - 1) for _ in range(100)]
    inputs += [random_value(H - 2, H + 1) for _ in range(60)]
    inputs += [random_value(H + 2, B - 2) for _ in range(80)]

    with mpmath.workprec(4 * P + 64):
        # around the threshold (2**-T), where Asinh switches to a ~ x
        for e in range(-T - 5, -T + 2):
            for s in range(2):
                inputs += [enc(1 << P, e, s), enc((1 << P) + 1, e, s), enc((2 << P) - 1, e, s)]
        # around a = 2**-8 and a = 1, where t = a + a**2/(1+sqrt(1+a**2)) reaches 2**-8 and u reaches 2
        two = mpmath.mpf(2)
        for center in (two ** -9, two ** -8, mpmath.mpf(1) - two ** -P, mpmath.mpf(1), mpmath.mpf(3) / 4, two):
            for d in range(-3, 4):
                for s in (1, -1):
                    inputs.append(nearest(s * (center + d * center * two ** (-P - 2))))
        # around a = 2**H, where Asinh switches to the log(2a) branch
        large = two ** H
        for d in range(-3, 4):
            for s in (1, -1):
                inputs.append(nearest(s * (large + d * large * two ** (-P - 2))))
        # near zero from both sides, where cancellation inside sqrt(1+a**2)
        # and the subsequent log1p is largest
        for e in range(-P - 4, -T - 2):
            for s in range(2):
                inputs += [enc(1 << P, e, s), enc((1 << P) + 1, e, s), enc((2 << P) - 1, e, s)]

    # the largest finite value, and its negation
    inputs.append(enc((1 << (P + 1)) - 1, B))
    inputs.append(enc((1 << (P + 1)) - 1, B, 1))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            with mpmath.workprec(4 * P + 64):
                x = dec(v)
            e = max(0, int(mpmath.floor(mpmath.log(abs(x), 2))) if x != 0 else 0)
            with mpmath.workprec(e + 4 * P + 64):
                y = round_bits(mpmath.asinh(x))
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


gen("asinh256", 236, 19, 256)
