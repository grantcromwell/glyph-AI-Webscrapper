#!/usr/bin/env python3
"""Argos Self-Update Spell — runs cerebrum clean and consolidate.
   Argos has built-in: cerebrum clean (prunes junk engrams).
   This runs that plus some housekeeping."""

import json, os, sys, subprocess, shutil

HOME = os.path.expanduser("~/Desktop/native")

def run_cerebrum(args, timeout=60):
    try:
        p = subprocess.run(
            ["go", "run", "./cmd/cerebrum"] + args,
            cwd=HOME, capture_output=True, text=True, timeout=timeout
        )
        return p.stdout, p.stderr, p.returncode
    except subprocess.TimeoutExpired:
        return "", "timeout", -1
    except FileNotFoundError:
        return "", "cerebrum binary missing", -1

def update():
    results = {}

    # 1. Run clean (prunes junk engrams)
    print("  Argos cleaning memory...")
    stdout, stderr, rc = run_cerebrum(["clean"])
    results["clean"] = {
        "rc": rc,
        "output": stdout.strip()[:200],
        "error": stderr.strip()[:200],
    }

    # 2. Check hipp campsize
    lore_path = os.path.join(HOME, "gut/lore")
    before_clean = 0
    after_clean = 0
    if os.path.exists(lore_path):
        try:
            before_clean = len([f for f in os.listdir(lore_path) if f.endswith('.svg')])
            after_clean = before_clean  # clean may not remove files
        except:
            pass

    results["engrams_before"] = before_clean

    print(json.dumps({"self": "argos", "action": "self_update", "results": results}))

if __name__ == "__main__":
    update()