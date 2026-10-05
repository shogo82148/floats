#!/usr/bin/env python3
# Generates testdata/cbrt128.txt.
# Each line contains the bits of x and the correctly rounded cbrt(x) in hexadecimal.
# The cube root is rounded with exact integer arithmetic.
#
# Usage: python3 scripts/gen_cbrt_testdata.py

import random


def icbrt(n):
    # floor(cbrt(n)) for n > 0
    x = 1 << ((n.bit_length() + 2) // 3)
    while True:
        y = (2 * x + n // (x * x)) // 3
        if y >= x:
            return x
        x = y


def gen(name, P, EB, seed):
    B = (1 << (EB - 1)) - 1
    width = (P + EB + 1) // 4
    rnd = random.Random(seed)

    def enc(m, e, s=0):
        # m is a normalized (P+1)-bit integer and the value is m × 2**(e-P), e >= 1-B
        return (s << (P + EB)) | ((e + B) << P) | (m - (1 << P))

    def decompose(v):
        # returns the sign, m, and e such that |x| = m × 2**(e-P) with m a normalized (P+1)-bit integer
        s = v >> (P + EB)
        e = (v >> P) & ((1 << EB) - 1)
        f = v & ((1 << P) - 1)
        if e == 0:
            l = f.bit_length()
            return s, f << (P + 1 - l), l - P - B
        return s, f | (1 << P), e - B

    def cbrt_bits(v):
        s, m, e = decompose(v)
        q, r = divmod(e, 3)
        # x = m × 2**(e-P) = (m × 2**r × 2**(2P)) × 2**(3q) × 2**(-3P), so that cbrt(x) = cbrt(n) × 2**(q-P)
        n = m << (r + 2 * P)
        R = icbrt(n)
        if 8 * n > (2 * R + 1) ** 3:
            R += 1
        if R >> (P + 1):
            R >>= 1
            q += 1
        return enc(R, q, s)

    def random_value(emin, emax, s=None):
        if s is None:
            s = rnd.getrandbits(1)
        return enc(rnd.getrandbits(P) | (1 << P), rnd.randint(emin, emax), s)

    inputs = []
    # random normal numbers
    inputs += [random_value(-B + 1, B) for _ in range(300)]
    inputs += [random_value(-20, 20) for _ in range(150)]
    # subnormal numbers
    for _ in range(60):
        f = rnd.getrandbits(rnd.randint(1, P)) or 1
        inputs.append(f | (rnd.getrandbits(1) << (P + EB)))
    # perfect cubes and their neighbors
    for _ in range(100):
        k = rnd.getrandbits(rnd.randint(1, (P + 1) // 3)) | 1
        j = rnd.randint(-(B // 3) + 40, B // 3 - 40)
        n = k ** 3
        l = n.bit_length()
        e = l - 1 + 3 * j
        m = n << (P + 1 - l)
        s = rnd.getrandbits(1)
        inputs.append(enc(m, e, s))
        inputs.append(enc(m, e, s) + 1)
        inputs.append(enc(m, e, s) - 1)
    # around the powers of two and eight, where the result is rounded up to the next power of two
    for e in (0, 1, 2, 3, 30, -30, 3 * (B // 3), -3 * (B // 3) + 3):
        for d in (-2, -1, 0, 1, 2):
            v = enc(1 << P, e) + d
            inputs += [v, v | (1 << (P + EB))]
    inputs.append(enc((1 << (P + 1)) - 1, 2))  # 8 - ulp
    inputs.append(enc((1 << (P + 1)) - 1, B))  # max finite
    inputs.append(enc((1 << (P + 1)) - 1, B) | (1 << (P + EB)))
    inputs.append(enc(1 << P, 1 - B))  # min normal
    inputs.append(1)  # min subnormal
    inputs.append(1 | (1 << (P + EB)))
    inputs.append(((1 << P) - 1))  # max subnormal

    # hard cases: the cube root is close to a midpoint of two floating-point numbers
    hard = []
    for r in range(3):
        for _ in range(60000):
            m = rnd.getrandbits(P) | (1 << P)
            g = icbrt((m << (r + 2 * P)) << 120)  # floor(cbrt(m × 2**r × 2**(2P)) × 2**40)
            frac = (g & ((1 << 40) - 1)) - (1 << 39)
            hard.append((abs(frac), m, r))
    hard.sort()
    for _, m, r in hard[:60]:
        e = 3 * rnd.randint(-100, 100) + r
        inputs.append(enc(m, e, rnd.getrandbits(1)))

    with open(f"testdata/{name}.txt", "w") as f:
        for v in inputs:
            f.write(f"{v:0{width}x} {cbrt_bits(v):0{width}x}\n")


gen("cbrt128", 112, 15, 128)
