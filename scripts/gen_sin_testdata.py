#!/usr/bin/env python3
# Generates testdata/sin128.txt and testdata/sin256.txt.
# Each line contains the bits of x, the correctly rounded sin(x),
# cos(x), and tan(x) in hexadecimal.
#
# Usage: python3 scripts/gen_sin_testdata.py

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
        x = mpmath.mpf((v & ((1 << P) - 1)) | (1 << P)) * mpmath.mpf(2) ** (e - B - P)
        return -x if s else x

    def round_bits(x):
        # round x to the nearest representable value. x must be normal.
        s = 1 if x < 0 else 0
        x = abs(x)
        e = int(mpmath.floor(mpmath.log(x, 2)))
        m = int(mpmath.nint(x * mpmath.mpf(2) ** (P - e)))
        if m >> (P + 1):
            m >>= 1
            e += 1
        return enc(m, e, s)

    def nearest(x):
        with mpmath.workprec(4 * P + 64):
            return round_bits(x)

    def random_value(emin, emax):
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), rnd.getrandbits(1))

    inputs = []
    # |x| <= Pi/4 and small arguments
    inputs += [random_value(-12, 1) for _ in range(150)]
    # tiny arguments
    inputs += [random_value(-B + 1, -13) for _ in range(20)]
    # moderate and large arguments
    inputs += [random_value(2, 200) for _ in range(60)]
    inputs += [random_value(201, B) for _ in range(20)]
    with mpmath.workprec(4 * P + 64):
        # near multiples of Pi (sin(x) is close to zero)
        for _ in range(40):
            nb = rnd.randint(1, P - 5)
            inputs.append(nearest((rnd.getrandbits(nb) | (1 << (nb - 1))) * mpmath.pi))
        # near odd multiples of Pi/2 (cos(x) is close to zero)
        for _ in range(40):
            nb = rnd.randint(1, P - 5)
            inputs.append(nearest((2 * (rnd.getrandbits(nb) | (1 << (nb - 1))) + 1) * mpmath.pi / 2))
        # near octant boundaries
        for _ in range(20):
            inputs.append(nearest(rnd.randint(1, 100) * mpmath.pi / 4))
    # the largest finite value
    inputs.append(enc((1 << (P + 1)) - 1, B))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            with mpmath.workprec(P + 64):
                x = dec(v)
            e = max(0, int(mpmath.floor(mpmath.log(abs(x), 2))))
            with mpmath.workprec(e + 4 * P + 64):
                s = round_bits(mpmath.sin(x))
                c = round_bits(mpmath.cos(x))
                t = round_bits(mpmath.tan(x))
            f.write(f"{v:0{width}x} {s:0{width}x} {c:0{width}x} {t:0{width}x}\n")


gen("sin128", 112, 15, 128)
gen("sin256", 236, 19, 256)
