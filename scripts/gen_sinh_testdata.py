#!/usr/bin/env python3
# Generates testdata/sinh128.txt, testdata/cosh128.txt, testdata/tanh128.txt, and testdata/sinh256.txt.
# Each line contains the bits of x and the correctly rounded sinh(x), cosh(x), or tanh(x) in hexadecimal.
#
# Usage: python3 scripts/gen_sinh_testdata.py

import random
import mpmath


def gen(names, P, EB, seed):
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
        if e > B:
            return (s << (P + EB)) | (((1 << EB) - 1) << P)  # infinity
        return enc(m, e, s)

    def nearest(x):
        with mpmath.workprec(4 * P + 64):
            return round_bits(x)

    def random_value(emin, emax, s=None):
        if s is None:
            s = rnd.getrandbits(1)
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    inputs = []
    # tiny arguments: sinh(x) ~ x
    inputs += [random_value(-B + 1, -P - 3) for _ in range(20)]
    inputs += [random_value(-P - 2, -P // 2) for _ in range(40)]
    # small arguments: |x| < ln(2)/128
    inputs += [random_value(-P // 2, -8) for _ in range(150)]
    # moderate arguments, where e**-x is not negligible
    inputs += [random_value(-7, 6) for _ in range(300)]
    # large arguments up to the overflow threshold
    inputs += [random_value(7, EB - 2) for _ in range(60)]
    with mpmath.workprec(4 * P + 64):
        ln2 = mpmath.log(2)
        # near multiples of ln(2)/128, where the reduced argument is
        # close to zero or to a boundary of the table
        for _ in range(60):
            m = rnd.choice([-1, 1]) * rnd.randint(1, 128 * 100)
            inputs.append(nearest(m * ln2 / 128))
        # near odd multiples of ln(2)/128 close to zero,
        # where the cancellation of e**x - e**-x is the largest
        for m in range(-15, 16, 2):
            for _ in range(5):
                inputs.append(nearest(m * ln2 / 128 * (1 + rnd.uniform(-2**-20, 2**-20))))
        # around the overflow threshold
        ovf = mpmath.log(mpmath.mpf(2) ** (B + 2) * (1 - mpmath.mpf(2) ** (-P - 2)))
        for s in (1, -1):
            inputs.append(nearest(s * ovf))
            for _ in range(10):
                inputs.append(nearest(s * (ovf - rnd.uniform(0, 1))))
    # the largest finite value
    inputs.append(enc((1 << (P + 1)) - 1, B))
    inputs.append(enc((1 << (P + 1)) - 1, B, 1))
    # around the threshold where sinh(x) rounds to x and cosh(x) rounds to 1
    for e in range(-P // 2 - 3, -P // 2 + 4):
        for s in range(2):
            inputs += [enc(1 << P, e, s), enc((1 << P) + 1, e, s), enc((2 << P) - 1, e, s)]

    # only for tanh: around the threshold where tanh(x) rounds to ±1
    tanh_inputs = list(inputs)
    with mpmath.workprec(4 * P + 64):
        one = mpmath.log(mpmath.mpf(2) ** (P + 3)) / 2
        for s in (1, -1):
            tanh_inputs.append(nearest(s * one))
            for _ in range(20):
                tanh_inputs.append(nearest(s * (one + rnd.uniform(-1, 1))))
    tanh_inputs += [random_value(6, 9) for _ in range(10)]

    for name, fn, xs in zip(names, (mpmath.sinh, mpmath.cosh, mpmath.tanh), (inputs, inputs, tanh_inputs)):
        with open(f"testdata/{name}.txt", "w") as f:
            for v in xs:
                with mpmath.workprec(P + 64):
                    x = dec(v)
                e = max(0, int(mpmath.floor(mpmath.log(abs(x), 2))))
                if fn is mpmath.tanh and abs(x) > 64:
                    y = round_bits(mpmath.sign(x))  # ±1
                elif abs(x) > 2 ** (EB - 1):
                    s = -1 if fn is mpmath.sinh and x < 0 else 1
                    y = round_bits(s * abs(x) * mpmath.mpf(2) ** B)  # ±Inf
                else:
                    with mpmath.workprec(e + 4 * P + 64):
                        y = round_bits(fn(x))
                f.write(f"{v:0{width}x} {y:0{width}x}\n")

gen(["sinh128", "cosh128", "tanh128"], 112, 15, 128)
gen(["sinh256"], 236, 19, 256)
