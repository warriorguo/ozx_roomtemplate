---
name: room-test
description: >
  Automated professional testing skill for OZX Room Template generate API.
  Use this skill when the user wants to test room generation, run generate API tests,
  validate room templates, check generation quality, or audit the roomtemplate backend.
  Trigger on: "test room generation", "run room tests", "test the generate API",
  "check bridge/fullroom/platform generation", "validate generated rooms", or any
  request to systematically test or audit the room template system.
---

# Room Test Skill

Professionally test the OZX Room Template generate API using equivalence partitioning,
boundary value analysis, and combinatorial testing. Validate every generated result,
file issues to MemoryFlow for anomalies, and record insights to md-editor.

**Target**: https://ozx-roomtpl.local.playquota.com/api/v1/generate/
**Issue ORT-23** | Map size fixed: **width=20, height=12**

---

## Overview

The testing process has 4 phases:

1. **Build test matrix** — select cases covering all important parameter combinations
2. **Execute in parallel** — use Agent tool to run test batches concurrently
3. **Validate results** — check each result against rules in `references/validation.md`
4. **Report & file issues** — deduplicate against MemoryFlow, file new bugs/requirements

---

## Phase 1: Test Matrix

Run ALL cases in Tier 1. Run Tier 2 if time permits or if Tier 1 reveals no issues.

### Tier 1 — Core Coverage (~18 cases)

**Fullroom** (POST /fullroom):

| # | doors | stageType | softEdgeCount | staticCount | railEnabled |
|---|-------|-----------|--------------|-------------|-------------|
| F1 | top,bottom | teaching | 3 | 3 | false |
| F2 | left,right | building | 3 | 3 | false |
| F3 | top,right | pressure | 5 | 5 | true |
| F4 | top,right,bottom | peak | 3 | 3 | false |
| F5 | top,right,bottom,left | boss | 0 | 0 | false |
| F6 | top,right,bottom,left | release | 5 | 8 | true |
| F7 | top | teaching | 0 | 0 | false |
| F8 | bottom,left | teaching | 8 | 8 | false |

**Bridge** (POST /bridge) — requires ≥2 doors:

| # | doors | stageType | softEdgeCount | staticCount | railEnabled |
|---|-------|-----------|--------------|-------------|-------------|
| B1 | left,right | teaching | 3 | 3 | false |
| B2 | top,bottom | building | 3 | 3 | true |
| B3 | top,right,bottom | teaching | 5 | 5 | false |
| B4 | top,right,bottom,left | building | 0 | 0 | true |
| B5 | left,right | (none/empty) | 3 | 3 | false |

**Platform** (POST /platform):

| # | doors | stageType | softEdgeCount | staticCount | railEnabled |
|---|-------|-----------|--------------|-------------|-------------|
| P1 | left,right | teaching | 3 | 3 | false |
| P2 | top,bottom | building | 3 | 3 | true |
| P3 | top,right | teaching | 0 | 0 | false |
| P4 | top,right,bottom,left | teaching | 5 | 5 | false |
| P5 | left | teaching | 3 | 3 | false |

### Tier 2 — Extended (run if Tier 1 passes cleanly)

Add: adjacent door pairs (top+right, right+bottom, bottom+left, left+top) for fullroom,
boss stage fullroom with full 4-door, platform with 3-door combos, bridge with adjacent
doors. Include zero-count enemies explicitly vs stage-derived counts.

---

## Phase 2: Parallel Execution

Divide Tier 1 into 3 batches and spawn agents simultaneously:

- **Batch A**: F1–F8 (fullroom)
- **Batch B**: B1–B5 (bridge)
- **Batch C**: P1–P5 (platform)

For each agent, provide:
1. The list of test cases to execute
2. The request format (see below)
3. Instructions to return a structured result (see below)

### Request Format

```bash
curl -sL -X POST "https://ozx-roomtpl.local.playquota.com/api/v1/generate/{roomType}" \
  -H "Content-Type: application/json" \
  -d '{
    "width": 20,
    "height": 12,
    "doors": ["top", "bottom"],
    "stageType": "teaching",
    "softEdgeCount": 3,
    "staticCount": 3,
    "railEnabled": false
  }'
```

- Omit `stageType` when testing without stage (empty string or omit key entirely)
- Valid door values: `"top"`, `"right"`, `"bottom"`, `"left"`
- On HTTP error, record the error and continue to next case

### Agent Instructions

Tell each batch agent:

> Execute the following test cases against the room generation API. For each case:
> 1. Make the HTTP request
> 2. If it fails (non-200 or error body), record: `{id, status: "error", error: "..."}`
> 3. If it succeeds, validate the payload using the rules in this skill's `references/validation.md`
> 4. Return a JSON array of results, one per test case, using this format:
>
> ```json
> [
>   {
>     "id": "F1",
>     "roomShape": "fullroom",
>     "params": { "doors": [...], "stageType": "...", "roomCategory": "normal", ... },
>     "status": "pass" | "fail" | "error",
>     "failures": ["description of each violation"],
>     "warnings": ["non-blocking quality observations"],
>     "bridgeCount": 0,
>     "entityCounts": { "chaser": 0, "zoner": 0, "dps": 0, "mobAir": 0, "static": 0 }
>   }
> ]
> ```
>
> Read `references/validation.md` for the full validation rules.

---

## Phase 3: Aggregate Results

After all agents complete, consolidate into a summary table:

```
ID | Room    | Doors           | Stage    | Status | Failures
F1 | full    | top,bottom      | teaching | PASS   |
B1 | bridge  | left,right      | teaching | FAIL   | bridge cell (14,4) floating
...
```

Count: total cases, passed, failed, errors.

---

## Phase 4: Issue Filing & Notes

### Filing Issues

For each unique failure pattern:

1. Use `memory-flow-pm` skill to search ORT project issues for similar bug/requirement
2. If already exists → note the existing issue key, skip creation
3. If new → file issue:
   - `bug` type for validation failures (P1 if structural, P2 if quality)
   - `requirement` type for missing features or improvement opportunities
   - Title should be specific: e.g. "Bridge tile floating: bottom cells lack ground contact in 4-row void"
   - Description: include the failing test case ID, params, exact failure message

### Recording Insights

After filing, use `md-editor` skill to append findings to topic
**"OZX RoomTemplate - 自动化测试方案设计"**:

Add a new section like:

```markdown
---

## Test Run — {date}

### Summary
- Ran N cases (fullroom: X, bridge: Y, platform: Z)
- Pass: N | Fail: N | Error: N

### Key Findings
- ...

### Issues Filed
- ORT-XX: ...

### Observations
- ...
```

---

## Quick Reference: Door Positions

For a 20×12 grid, door center positions:
- **top**: (10, 0)
- **bottom**: (10, 11)
- **left**: (0, 6)
- **right**: (19, 6)

Forbidden zone for entities: Manhattan distance ≤ 2 from door center.

---

## Handling Edge Cases

- **Bridge with 1 door**: expect 400 error — this is correct behavior, not a bug
- **Peak stage on bridge**: stage rules prohibit this — if API returns error, that's correct
- **Release stage**: minimal or no enemies expected — validate structure only
- **Boss stage**: validate 6×6 center clear zone specifically

See `references/validation.md` for full rules.
