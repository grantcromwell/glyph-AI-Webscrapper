#!/usr/bin/env python3
"""Argos Self-Analysis Spell — queries the existing cerebrum diagnostics.
   Argos already has: stats, coverage, clean, train, foresee.
   This wraps them into a standard self-analysis JSON output."""

import json, os, sys, subprocess

HOME = os.path.expanduser("~/Desktop/native")

def run_cerebrum(args):
    try:
        p = subprocess.run(
            ["go", "run", "./cmd/cerebrum"] + args,
            cwd=HOME, capture_output=True, text=True, timeout=30
        )
        return p.stdout, p.stderr, p.returncode
    except subprocess.TimeoutExpired:
        return "", "timeout", -1
    except FileNotFoundError:
        return "", "cerebrum not found (needs go run)", -1

def analyze():
    # 1. Run stats
    stdout, stderr, rc = run_cerebrum(["stats"])
    stats_result = {}
    if rc == 0:
        # Parse stats output — argos stats are human-readable
        for line in stdout.split('\n'):
            line = line.strip()
            if ':' in line:
                k, v = line.split(':', 1)
                stats_result[k.strip().lower().replace(' ', '_')] = v.strip()

    # 2. Run coverage
    cov_stdout, _, cov_rc = run_cerebrum(["coverage"])
    coverage = 0.0
    if cov_rc == 0:
        for line in cov_stdout.split('\n'):
            if '%' in line and 'coverage' in line.lower():
                try:
                    coverage = float(line.split('%')[0].split()[-1]) / 100.0
                except:
                    pass

    # 3. Check hippocampus files
    hippocampus_path = os.path.join(HOME, "bin/hippocampus")
    engram_count = 0
    if os.path.exists(hippocampus_path):
        try:
            p = subprocess.run(["ls", os.path.join(HOME, "gut/lore/")],
                               capture_output=True, text=True, timeout=5)
            engram_count = len([f for f in p.stdout.split('\n') if f.endswith('.svg')])
        except:
            pass

    # 4. Check myelin files for learned weights
    myelin_path = os.path.join(HOME, "myelin")
    myelin_files = {}
    if os.path.exists(myelin_path):
        for f in os.listdir(myelin_path):
            if f.endswith('.svg'):
                fp = os.path.join(myelin_path, f)
                myelin_files[f] = os.path.getsize(fp)

    # 5. Check lexicon
    lexicon = {}
    lexicon_path = os.path.join(myelin_path, "lexicon.svg")
    if os.path.exists(lexicon_path):
        try:
            with open(lexicon_path) as f:
                content = f.read()
            # Count logograms
            import re
            logograms = re.findall(r'logogram', content)
            lexicon['logograms'] = len(logograms)
        except:
            lexicon['logograms'] = 0

    health = 0.5
    if coverage > 0:
        health += coverage * 0.3
    if engram_count > 10:
        health += 0.1
    if 'cerebellum' in str(myelin_files):
        health += 0.1
    health = max(0, min(1, health))

    output = {
        "self": "argos",
        "coverage": round(coverage, 3),
        "engrams": engram_count,
        "myelin_files": list(myelin_files.keys()),
        "lexicon_logograms": lexicon.get('logograms', 0),
        "health": round(health, 3),
        "needs_clean": health < 0.4,
        "stats": stats_result,
        "recommendation": "healthy" if health > 0.6 else ("needs_attention" if health > 0.3 else "stale"),
    }
    print(json.dumps(output))

if __name__ == "__main__":
    analyze()