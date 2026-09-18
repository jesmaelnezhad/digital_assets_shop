#!/usr/bin/env python3
"""Copy every tag from the host Docker registry into the in-cluster HTTPS registry."""
import json
import os
import subprocess
import sys
import urllib.request

OLD = os.environ.get("OLD_REGISTRY", "127.0.0.1:30099")
NEW = os.environ["REG_HOST"]
LOG_PATH = os.environ.get("MIRROR_LOG", "/tmp/registry-mirror.log")


def catalog():
    with urllib.request.urlopen("http://%s/v2/_catalog?n=1000" % OLD) as r:
        return json.load(r).get("repositories") or []


def tags(repo):
    with urllib.request.urlopen("http://%s/v2/%s/tags/list" % (OLD, repo)) as r:
        return json.load(r).get("tags") or []


def run(cmd, log):
    p = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    log.write("$ %s\n%s\n" % (" ".join(cmd), p.stdout))
    log.flush()
    return p.returncode, p.stdout


def main():
    ok = 0
    fail = 0
    with open(LOG_PATH, "a") as log:
        repos = catalog()
        print("repos=%d" % len(repos), flush=True)
        for repo in repos:
            ts = tags(repo)
            print("-- %s tags=%s" % (repo, ts), flush=True)
            for tag in ts:
                src = "%s/%s:%s" % (OLD, repo, tag)
                dst = "%s/%s:%s" % (NEW, repo, tag)
                rc, _ = run(["docker", "pull", src], log)
                if rc != 0:
                    print("FAIL pull %s" % src, flush=True)
                    fail += 1
                    continue
                rc, _ = run(["docker", "tag", src, dst], log)
                if rc != 0:
                    print("FAIL tag %s" % dst, flush=True)
                    fail += 1
                    continue
                rc, _ = run(["docker", "push", dst], log)
                if rc != 0:
                    print("FAIL push %s" % dst, flush=True)
                    fail += 1
                    continue
                print("OK %s" % dst, flush=True)
                ok += 1
    print("MIRROR_SUMMARY ok=%d fail=%d" % (ok, fail), flush=True)
    return 1 if fail else 0


if __name__ == "__main__":
    sys.exit(main())
