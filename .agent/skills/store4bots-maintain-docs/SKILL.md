---
name: store4bots-maintain-docs
description: Keep Store4bots docs, skills, rules, and scripts accurate to the live system. Use after any infra, API, UI, or process change.
---

# Maintain docs, skills, scripts, rules

The repo must describe **what is running now** and **how to recreate it**. Not past outages, not abandoned approaches.

## After every material change

1. Update the one identity file if hosts/ports changed: `config/site.env`.
2. Update the skill that owns the topic (kubernetes, network, databases, build-deploy, …).
3. Update the matching `docs/*.md` current-state page.
4. Update scripts if a command or flag changed.
5. Grep for leftover IPs or domains that belong only in `config/site.env`, and for prose that describes a past approach instead of the current one.

## What to keep vs drop

Keep: reusable procedures (how to migrate RED to a new VM; how to rotate registry passwords).

Drop: diaries of past incidents, old IPs in prose, and “do not do X” notes that only make sense if you know an abandoned design.

Write the current design in the affirmative. Do not name discarded schemes.

## Identity rule

Public hosts, ports, and image repo name live in `config/site.env`. Secrets in `config/site.secrets.env` (gitignored). Docs and skills use `$STAGING_HOST`, `$RED_HOST`, `{{REGISTRY_HOST}}` — never scatter this install’s domain or IP.

## Skills

Edit only `.agent/skills/<name>/SKILL.md`. That is the skills tree for every agent.

## Tests report

Pass counts only in `tests/REPORT.md`. Do not paste “N tests passed” into skills or architecture docs.
