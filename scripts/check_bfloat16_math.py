#!/usr/bin/env python3
# Checks the mathematical functions of BFloat16 for all the 65536 inputs.
# The results of "go run ./internal/cmd/bf16dump" are compared with the correctly rounded results of mpmath.
#
# The inputs whose results are NaN, infinite or complex in mpmath are not compared;
# TestBFloat16_MathNonFinite in bfloat16_math_test.go checks them.
#
# Usage: python3 scripts/check_bfloat16_math.py [-j JOBS] [-d DIR] [NAME ...]
#
# NAME is a name of a job (e.g. Exp, Jn2, Pow3, Atan2_8, Hypot0, LgammaSign).
# All the jobs are checked if no NAME is given. It takes about 10 minutes.
# DIR is the directory for the output of bf16dump. It is a temporary directory by default.
# If DIR already has the output of all the selected jobs, bf16dump is not run again.

import argparse
import math
import os
import struct
import subprocess
import sys
import tempfile
from multiprocessing import Pool

import mpmath
from mpmath import mp, mpf

# the values of the second argument of Pow, Atan2 and Hypot. The same as internal/cmd/bf16dump/main.go.
SECOND_ARGS = [
    -3, -2, -1, -0.5, 0.5, 1.0 / 3, 0.25, 1.5, 2, 3, 5, 10, -10, 0.1, 100, 1000, 0.001, 2.5, 7, -0.7,
    0, 1, -1, 0.75, 1e5, 1e-5, 3e38, 1e-40,
]

UNARY = [
    "Exp", "Exp2", "Expm1", "Log", "Log2", "Log10", "Log1p",
    "Sin", "Cos", "Tan", "Asin", "Acos", "Atan",
    "Sinh", "Cosh", "Tanh", "Asinh", "Acosh", "Atanh",
    "Cbrt", "Gamma", "Erf", "Erfc", "Erfinv", "Erfcinv",
    "J0", "J1", "Y0", "Y1", "Jn1", "Jn-1", "Jn2", "Jn5", "Yn2", "Yn5",
    "Lgamma", "LgammaSign",
]

def bf16_value(bits):
    """returns the value of the BFloat16 with the bits as a float."""
    return struct.unpack(">f", struct.pack(">I", bits << 16))[0]


def round_bf16(r):
    """rounds the finite mpf r to BFloat16, ties to even. returns the bits and whether r is a tie."""
    if r == 0:
        return 0, False
    sign = 0x8000 if r < 0 else 0
    _, man, exp, bc = r._mpf_
    # |r| = man * 2**exp
    e = exp + bc - 1
    if e >= 200:
        return sign | 0x7F80, False
    if e < -300:
        return sign, False
    # the exponent of the ULP. The subnormal numbers have the fixed one.
    ulp = -133 if e < -126 else e - 7
    k = ulp - exp
    tie = False
    if k <= 0:
        n = man << -k
    else:
        n = man >> k
        rem = man & ((1 << k) - 1)
        half = 1 << (k - 1)
        if rem > half or (rem == half and n & 1):
            n += 1
        tie = rem == half
    if n == 0:
        return sign, tie
    if n.bit_length() - 1 + ulp >= 128:
        return sign | 0x7F80, tie
    v = math.ldexp(n, ulp)  # exactly representable in float32
    return sign | struct.unpack(">I", struct.pack(">f", v))[0] >> 16, tie


def evaluate(name, param, x):
    """returns the exact function value as mpf. It returns None if the value is not compared."""
    X = mpf(x)
    if name == "Exp":
        return mpmath.exp(X)
    if name == "Exp2":
        return mpf(2) ** X
    if name == "Expm1":
        return mpmath.expm1(X)
    if name == "Log":
        return mpmath.log(X)
    if name == "Log2":
        return mpmath.log(X, 2)
    if name == "Log10":
        return mpmath.log10(X)
    if name == "Log1p":
        return mpmath.log1p(X)
    if name == "Sin":
        return mpmath.sin(X)
    if name == "Cos":
        return mpmath.cos(X)
    if name == "Tan":
        return mpmath.tan(X)
    if name == "Asin":
        return mpmath.asin(X)
    if name == "Acos":
        return mpmath.acos(X)
    if name == "Atan":
        return mpmath.atan(X)
    if name == "Sinh":
        return mpmath.sinh(X)
    if name == "Cosh":
        return mpmath.cosh(X)
    if name == "Tanh":
        return mpmath.tanh(X)
    if name == "Asinh":
        return mpmath.asinh(X)
    if name == "Acosh":
        return mpmath.acosh(X)
    if name == "Atanh":
        return mpmath.atanh(X)
    if name == "Cbrt":
        return mpmath.cbrt(X)
    if name == "Gamma":
        return mpmath.gamma(X)
    if name == "Erf":
        return mpmath.erf(X)
    if name == "Erfc":
        return mpmath.erfc(X)
    if name == "Erfinv":
        return mpmath.erfinv(X)
    if name == "Erfcinv":
        return mpmath.erfinv(1 - X)
    if name == "J0":
        return mpmath.besselj(0, X)
    if name == "J1" or name == "Jn1":
        return mpmath.besselj(1, X)
    if name == "Jn-1":
        return -mpmath.besselj(1, X)
    if name == "Y0":
        return mpmath.bessely(0, X)
    if name == "Y1":
        return mpmath.bessely(1, X)
    if name == "Jn2":
        return mpmath.besselj(2, X)
    if name == "Jn5":
        return mpmath.besselj(5, X)
    if name == "Yn2":
        return mpmath.bessely(2, X)
    if name == "Yn5":
        return mpmath.bessely(5, X)
    if name == "Lgamma":
        return mpmath.log(abs(mpmath.gamma(X)))
    if name == "LgammaSign":
        # the sign of Gamma, except for the poles and the zero
        if x <= 0 and x == math.floor(x):
            return None
        return mpf(1) if mpmath.gamma(X) > 0 else mpf(-1)
    if name == "Pow":
        return mpmath.power(X, mpf(param))
    if name == "Atan2":
        return mpmath.atan2(X, mpf(param))
    if name == "Hypot":
        return mpmath.hypot(X, mpf(param))
    raise ValueError(name)


def precision(name):
    # 1 - X must keep the digits of the small X.
    return 700 if name == "Erfcinv" else 250


# the functions that overflow or underflow trivially for a large argument
HUGE = {
    ("Exp", 1): 0x7F80, ("Exp", -1): 0x0000,
    ("Exp2", 1): 0x7F80, ("Exp2", -1): 0x0000,
    ("Erfc", 1): 0x0000, ("Erfc", -1): 0x4000,
}


def check(job):
    """checks a job, and returns the name, the number of mismatches, the number of skipped inputs, and examples."""
    directory, fname, name, param = job
    with open(os.path.join(directory, fname + ".txt")) as f:
        got = [int(line, 16) for line in f]
    assert len(got) == 1 << 16, fname
    bad = []
    skipped = 0
    for i in range(1 << 16):
        x = bf16_value(i)
        if x != x or abs(x) == math.inf:
            skipped += 1
            continue
        if name == "Atan2" and i == 0x8000:
            skipped += 1  # mpf cannot represent -0
            continue
        if (name, 1 if x > 0 else -1) in HUGE and abs(x) > 1e4:
            want = HUGE[(name, 1 if x > 0 else -1)]
            if got[i] != want:
                bad.append((i, x, got[i], want))
            continue

        mp.prec = precision(name)
        try:
            r = evaluate(name, param, x)
        except (ValueError, ZeroDivisionError, OverflowError):
            skipped += 1  # poles and the domain errors
            continue
        if r is None or isinstance(r, mpmath.mpc) or not mpmath.isfinite(r):
            skipped += 1
            continue

        if name == "LgammaSign":
            want = 0x0001 if r > 0 else 0xFFFF
            tie = False
        else:
            want, tie = round_bf16(r)
            if tie:
                # check that it is a real tie, not a rounding error of the working precision.
                mp.prec = precision(name) * 3
                want, tie = round_bf16(evaluate(name, param, x))
        if got[i] != want and not (got[i] & 0x7FFF == 0 and want == 0):
            bad.append((i, x, got[i], want))
    return fname, len(bad), skipped, bad[:5]


def main():
    parser = argparse.ArgumentParser(description="Checks the mathematical functions of BFloat16 with mpmath.")
    parser.add_argument("-j", "--jobs", type=int, default=os.cpu_count())
    parser.add_argument("-d", "--dir", default=None)
    parser.add_argument("names", nargs="*")
    args = parser.parse_args()

    jobs = [(n, n, None) for n in UNARY]
    mp.prec = 100
    for k, y in enumerate(SECOND_ARGS):
        # the second argument is rounded to BFloat16 in bf16dump
        y = bf16_value(round_bf16(mpf(y))[0])
        jobs.append((f"Pow{k}", "Pow", y))
        jobs.append((f"Atan2_{k}", "Atan2", y))
        jobs.append((f"Hypot{k}", "Hypot", y))
    if args.names:
        names = set(args.names)
        unknown = names - {j[0] for j in jobs}
        if unknown:
            parser.error("unknown names: " + ", ".join(sorted(unknown)))
        jobs = [j for j in jobs if j[0] in names]

    with tempfile.TemporaryDirectory() as tmp:
        directory = args.dir or tmp
        # j[0] is the name of the output file of the job.
        if not all(os.path.exists(os.path.join(directory, j[0] + ".txt")) for j in jobs):
            subprocess.run(["go", "run", "./internal/cmd/bf16dump", directory], check=True)
        jobs = [(directory, *job) for job in jobs]

        print(len(jobs), "jobs", flush=True)
        failed = 0
        with Pool(args.jobs) as pool:
            for fname, nbad, skipped, bad in pool.imap_unordered(check, jobs):
                status = "ok" if nbad == 0 else "FAIL"
                print(f"{status} {fname}: mismatches={nbad} skipped={skipped}", flush=True)
                for i, x, got, want in bad:
                    print(f"  input={i:#06x} ({x!r}) got={got:#06x} want={want:#06x}")
                if nbad:
                    failed += 1
    if failed:
        print(failed, "jobs failed")
        sys.exit(1)
    print("all jobs passed")


if __name__ == "__main__":
    main()
