#!/usr/bin/env python3
# Generates testdata/sin32.txt, testdata/sin32_hard.txt, testdata/sin128.txt, and testdata/sin256.txt.
# Each line contains the bits of x, the correctly rounded sin(x),
# cos(x), and tan(x) in hexadecimal.
#
# Usage: python3 scripts/gen_sin_testdata.py

import random
import mpmath


def gen(name, P, EB, seed, emid=200, hard=None):
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

    if hard is not None:
        # hard-to-round cases only
        inputs = list(hard)
    else:
        inputs = []
        # |x| <= Pi/4 and small arguments
        inputs += [random_value(-12, 1) for _ in range(150)]
        # tiny arguments
        inputs += [random_value(-B + 1, -13) for _ in range(20)]
        # moderate and large arguments
        inputs += [random_value(2, emid) for _ in range(60)]
        inputs += [random_value(emid + 1, B) for _ in range(20)]
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


# Float32 inputs whose sin or cos is very close to the midpoint of two adjacent Float32 values.
# They are found by checking all Float32 values with math.Sin and math.Cos in float64.
hard32 = [
    0x39800000, 0x3a544395, 0x3c107fe6, 0x3dcf5597, 0x3ef3830f, 0x42378db8,
    0x424790ce, 0x4371ade3, 0x45a8abb3, 0x46199998, 0x47a0e238, 0x4967cb9b,
    0x4986afee, 0x4a01dca4, 0x4aa5a796, 0x4fb56937, 0x521945ed, 0x52d9d3fe,
    0x543f6e04, 0x55cafb2a, 0x55e5235d, 0x58dfb085, 0x5922aa80, 0x59443c0a,
    0x5a8c921b, 0x5a935f4c, 0x5dadd689, 0x5f18b878, 0x5f208d82, 0x6115cb11,
    0x616d8730, 0x61703976, 0x61dfc847, 0x6446cec0, 0x653cee8f, 0x67a9242b,
    0x6a3f60ff, 0x6d734599, 0x73243f06, 0x76d7173f, 0x7908cd73, 0x79d1f6d3,
    0x7a38ab34, 0x7a4b1a27, 0x7a5aacdb, 0x7a817b08, 0x7c64841e, 0x7c69ae1e,
    0xb9800000, 0xba544395, 0xbc107fe6, 0xbdcf5597, 0xbef3830f, 0xc2378db8,
    0xc24790ce, 0xc371ade3, 0xc5a8abb3, 0xc6199998, 0xc7a0e238, 0xc967cb9b,
    0xc986afee, 0xca01dca4, 0xcaa5a796, 0xcfb56937, 0xd21945ed, 0xd2d9d3fe,
    0xd43f6e04, 0xd5cafb2a, 0xd5e5235d, 0xd8dfb085, 0xd922aa80, 0xd9443c0a,
    0xda8c921b, 0xda935f4c, 0xddadd689, 0xdf18b878, 0xdf208d82, 0xe115cb11,
    0xe16d8730, 0xe1703976, 0xe1dfc847, 0xe446cec0, 0xe53cee8f, 0xe7a9242b,
    0xea3f60ff, 0xed734599, 0xf3243f06, 0xf6d7173f, 0xf908cd73, 0xf9d1f6d3,
    0xfa38ab34, 0xfa4b1a27, 0xfa5aacdb, 0xfa817b08, 0xfc64841e, 0xfc69ae1e,
]

gen("sin32", 23, 8, 32, emid=60)
gen("sin32_hard", 23, 8, 32, hard=hard32)
gen("sin128", 112, 15, 128)
gen("sin256", 236, 19, 256)
