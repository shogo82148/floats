#!/usr/bin/env python3
# Generates testdata/exp128.txt and testdata/exp2_128.txt.
# Each line contains the bits of x and the correctly rounded exp(x) or 2**x in hexadecimal.
#
# Usage: python3 scripts/gen_exp_testdata.py

import random
import mpmath


def gen(name, P, EB, seed, base2=False):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)
    inf = ((1 << EB) - 1) << P

    def enc(m, e, s=0):
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def dec(v):
        s = v >> (P + EB)
        e = (v >> P) & ((1 << EB) - 1)
        x = mpmath.mpf((v & ((1 << P) - 1)) | (1 << P)) * mpmath.mpf(2) ** (e - B - P)
        return -x if s else x

    def round_bits(x):
        # round the positive value x to the nearest representable value.
        e = max(int(mpmath.floor(mpmath.log(x, 2))), 1 - B)
        m = int(mpmath.nint(x * mpmath.mpf(2) ** (P - e)))
        if m >> (P + 1):
            m >>= 1
            e += 1
        if e > B:
            return inf
        if m >> P == 0:
            # subnormal (or the smallest normal number after rounding up)
            return m
        return enc(m, e)

    def nearest(x):
        with mpmath.workprec(4 * P + 64):
            s = 1 if x < 0 else 0
            v = round_bits(abs(x))
            return v | (s << (P + EB))

    def random_value(emin, emax, s=None):
        if s is None:
            s = rnd.getrandbits(1)
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    inputs = []
    # tiny arguments: exp(x) ~ 1 + x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(20)]
    inputs += [random_value(-P - 2, -8) for _ in range(150)]
    # tiny arguments where 1 + x is close to a midpoint
    for e in range(-P - 2, -P + 1):
        for d in range(4):
            inputs += [enc((1 << P) + d, e, 0), enc((1 << P) + d, e, 1), enc((2 << P) - 1 - d, e, 0), enc((2 << P) - 1 - d, e, 1)]
    # moderate arguments
    inputs += [random_value(-7, 6) for _ in range(250)]
    # large arguments
    inputs += [random_value(7, 13) for _ in range(80)]
    with mpmath.workprec(4 * P + 64):
        # ln2 is the argument that doubles the result.
        ln2 = mpmath.mpf(1) if base2 else mpmath.log(2)
        if base2:
            # exact results, and exactly representable multiples of 1/64
            for _ in range(40):
                inputs.append(nearest(mpmath.mpf(rnd.randint(-B - P - 2, B + 1))))
            for _ in range(40):
                inputs.append(nearest(mpmath.mpf(rnd.randint(-64 * (B + P + 2), 64 * (B + 1))) / 64))
            for k in range(-B - P - 2, -B - P + 3):
                inputs.append(nearest(mpmath.mpf(k)))
                inputs.append(nearest(mpmath.mpf(k) + mpmath.mpf(1) / 2))
            for k in (-1, 1, 2, B, B + 1):
                inputs.append(nearest(mpmath.mpf(k)))
        # near multiples of ln(2)/128, where the reduced argument is
        # close to zero or to a boundary of the table
        for _ in range(60):
            m = rnd.choice([-1, 1]) * rnd.randint(1, 128 * 16000)
            inputs.append(nearest(m * ln2 / 128))
        for m in range(-15, 16, 2):
            inputs.append(nearest(m * ln2 / 128 * (1 + rnd.uniform(-2**-20, 2**-20))))
        # around the overflow threshold
        ovf = mpmath.log(mpmath.mpf(2) ** (B + 1) * (1 - mpmath.mpf(2) ** (-P - 2)))
        inputs.append(nearest(ovf))
        for _ in range(10):
            inputs.append(nearest(ovf - rnd.uniform(0, 1)))
        # subnormal results and around the underflow threshold
        emin = (1 - B) * ln2
        # rounded up to the smallest normal number
        inputs.append(nearest(emin))
        inputs.append(nearest(emin) + 1)
        for _ in range(40):
            inputs.append(nearest(emin - rnd.uniform(0, P + 1) * ln2))
        unf = (1 - B - P - 1) * ln2
        inputs.append(nearest(unf))
        for _ in range(5):
            inputs.append(nearest(unf + rnd.uniform(0, 2**-10)))
    # the largest finite value
    inputs.append(enc((1 << (P + 1)) - 1, B))
    inputs.append(enc((1 << (P + 1)) - 1, B, 1))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            with mpmath.workprec(P + 64):
                x = dec(v)
            if x > 2 ** (EB - 1):
                y = inf
            elif x < -(B + P + 3):
                y = 0
            else:
                e = max(0, int(mpmath.floor(mpmath.log(abs(x), 2))))
                with mpmath.workprec(e + 4 * P + 64):
                    y = round_bits(mpmath.power(2, x) if base2 else mpmath.exp(x))
            f.write(f"{v:0{width}x} {y:0{width}x}\n")


gen("exp128", 112, 15, 128)
gen("exp2_128", 112, 15, 129, base2=True)
