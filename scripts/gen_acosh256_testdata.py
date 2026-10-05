#!/usr/bin/env python3
# Generates testdata/acosh256.txt.
# Each line contains the bits of x and the correctly rounded acosh(x) in hexadecimal.
#
# Usage: python3 scripts/gen_acosh256_testdata.py

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

    # Acosh approximates log(2a) for a >= 2**H.
    H = 128

    inputs = []
    # arguments close to 1, where a**2-1 cancels and acosh(a) ~ sqrt(2(a-1))
    inputs += [enc((1 << P) + d, 0, 0) for d in range(1, 9)]
    inputs += [enc((1 << P) + rnd.getrandbits(rnd.randint(1, P)) + 1, 0, 0) for _ in range(200)]
    inputs += [enc((1 << P) + (1 << rnd.randint(0, P - 1)), 0, 0) for _ in range(40)]
    inputs += [enc((2 << P) - d, 0, 0) for d in range(1, 9)]
    # 1 <= a < 2 across the point where t = (a-1) + sqrt(a**2-1) reaches 2**-8, and 2 <= a < 2**H
    inputs += [random_value(0, 0, 0) for _ in range(150)]
    inputs += [random_value(1, 1, 0) for _ in range(100)]
    inputs += [random_value(2, 40, 0) for _ in range(100)]
    inputs += [random_value(41, H - 1, 0) for _ in range(100)]
    inputs += [random_value(H - 2, H + 1, 0) for _ in range(60)]
    inputs += [random_value(H + 2, B - 2, 0) for _ in range(80)]

    with mpmath.workprec(4 * P + 64):
        two = mpmath.mpf(2)
        # around t = 2**-8 (a - 1 ~ 2**-17), where the reduction starts
        for e in (-17, -16):
            for d in range(-3, 4):
                inputs.append(nearest(1 + two ** e + d * two ** (-P)))
        # around a = 2.0 and 4.0
        for center in (two, two * 2):
            for d in range(-3, 4):
                inputs.append(nearest(center + d * center * two ** (-P - 1)))
        # around a = 2**H, where Acosh switches to the log(2a) branch
        large = two ** H
        for d in range(-3, 4):
            inputs.append(nearest(large + d * large * two ** (-P - 1)))

    # the largest finite value
    inputs.append(enc((1 << (P + 1)) - 1, B))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            with mpmath.workprec(4 * P + 64):
                x = dec(v)
            e = max(0, int(mpmath.floor(mpmath.log(abs(x), 2))) if x != 0 else 0)
            with mpmath.workprec(e + 4 * P + 64):
                y = round_bits(mpmath.acosh(x))
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


gen("acosh256", 236, 19, 256)
