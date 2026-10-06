# roombatch — batch room generation with acceptance rules

Generates rooms through the running editor backend and keeps only the ones that
pass the acceptance rules agreed in the 2026-09-14 play-test review.

```bash
# the backend must be running; override with ROOM_API if it is not on :8099
cd tile-backend && ./bin/ozx-roomeditor --port 8099 --no-browser &

python3 tools/roombatch/batch.py                       # dry run: plan + verdicts
python3 tools/roombatch/batch.py --save                # write the accepted rooms
python3 tools/roombatch/batch.py --target 3 --attempts 30 --stages teaching,building
python3 tools/roombatch/batch.py --size 28x10 --category test --save
```

`batch.py` tops up every `stage x doors` category that holds fewer than
`--target` templates, regenerating until a room passes the HARD rules or
`--attempts` runs out (then it keeps the least-bad attempt and says so).
`start` and `boss` are skipped — they field no enemies, so there is nothing to
judge. Rooms are written through `POST /templates`, so `fsstore` derives the
filename and the door rotation exactly as the editor would.

## HARD rules — a room that trips one is thrown away and regenerated

Nine rules. Each one is either an explicit instruction from review, or something
every template already in the project satisfies. `judge.py` is the authority;
this table is what it implements.

| # | rule | where it came from |
|---|------|--------------------|
| 1 | every zoner spawn is a full 2×2 block | review: the ORT-103 1×1 fallback was removed from the generator |
| 2 | no zoner in the room's visual bottom third | review: the player walks in from there and fires automatically |
| 3 | no two spawns of the same category 8-adjacent (chaser, dps) | ORT-93 — the game collapses them and crashes |
| 4 | every carved void is a clean rectangle (bbox fill 1.0) | review: the four rooms accepted by eye had every void at 1.0, the one rejected as "fragmented" had two at 0.67 |
| 5 | every carved void is at least 2 cells thick in both directions | review: a 2×1 nick at a corner reads as a mistake |
| 6 | no 1-cell ground spurs | all 23 hand-authored rooms have zero |
| 7 | no unreachable ground island | a slab the player cannot get to, sometimes carrying spawns |
| 8 | teaching/building/release floor fill ≤ 0.9 | review: those stages want holes in the floor, and carving is a dice roll |
| 9 | with ≥12 ground spawns: no entity-free quadrant, and no half of the room carrying >1.8× the other; with ≥4 static blocks: cover in at least 3 quadrants | review: rooms with a bare half or all the cover in one corner |

Rule 9's threshold is calibrated, not guessed: every room accepted by eye sat at
or below 1.63 on the half-to-half ratio, and the two rejected as one-sided were
at 2.4 (left/right) and 1.9 (top/bottom). Quadrant ratios do not separate those
cases; halves do.

## Rail rooms (ORT-138)

Three rules come from the rail chapter, all in `judge.py`:

- **rail length 35-45 cells** when the layer is non-empty — long enough for the
  cart to be worth boarding, short enough not to turn the floor into a grid.
  Matches the four shipped rail rooms (35/36/40/44).
- **every open door needs a 2x2 walkable, static-free passage.** One unit of
  clearance funnels the player and the cart through a single tile; 17% of rooms
  failed this before it was a rule.
- **half-density gate lowered from 12 spawns to 10** — two rail rooms at 10 and
  11 spawns came out 9:1 and 8:3 and had to be redone by eye.

`--diverse` also compares the `rail` layer (limit 0.60): the network is derived
from the ground shape, so two rooms with similar floors get *identical* track
(measured max 1.00 over ten same-stage rooms). Note that two rooms with no rail
at all are not "identical" — `_jaccard` returns 0 for two empty sets, or every
non-rail pair in the library would be reported as a collision.

On the generator side, mobAir now avoids rail cells (`isValidMobAirPositionNew`),
with a final top-up pass that *allows* the overlap once clear cells run out:
staying off the track is a preference, but a room's slot counts have to stay in
its family's range or OZX's LevelPlanValidator sees less intake capacity than the
plan expects.

## Layout diversity (`--diverse`, ORT-132)

Rooms sharing a `(stageType, openDoors)` family share the candidate pool the
level generator draws from, so two similar rooms in one family read as a repeat
when they land next to each other. `--diverse` rejects a candidate whose layout
is too close to any room already in its family (and to the others produced in
the same run), by Jaccard overlap of cell sets:

| layer | limit | why this number |
|-------|------:|-----------------|
| ground | 0.95 | pathing. Two shipped peak d5 rooms have *identical* ground (1.00) — that is the thing ORT-132 was filed about, so the bar here is deliberately stricter than what ships |
| static | 0.35 | cover position. Shipped d5 pairs run 0.07-0.50 |
| enemies (all four layers unioned) | 0.35 | shipped d5 pairs run 0.11-0.43 |

The check is **incremental** — a room is compared against the family as it stood
when that room was generated — so a later sibling can still end up close to an
earlier one. With `--save` the run therefore ends with a whole-family re-check
that names any offending pair; **redo those rooms** (delete the file, rerun that
cell) rather than leaving a near-duplicate in the pool. Three of ORT-132's first
eleven rooms needed exactly that.

A dense family can saturate: peak d5 fields 40+ spawn cells in a 160-cell room,
and after four rooms a fifth could not get below 0.41 enemy overlap in 120
attempts. Four distinct rooms beat five with two near-duplicates — the pool size
is a means, layout difference is the goal.

**Zoner block spacing is deliberately NOT a hard rule.** The generator prefers
2 clear cells between blocks and falls back to merely not touching — asked
about it in review, the answer was "not necessarily", so a room is never
rejected for it. `judge.py` reports it as a soft note instead.

Hard-rule pass rate at 16×10 with four doors, 25 rooms per stage: teaching
12/25, building 12/25, pressure 22/25, peak 12/25, release 14/25 — so roughly
2 attempts per accepted room, well inside the default `--attempts 20`. The
dominant failures are non-rectangular voids (27) and ground spurs (21), both
left over from `ensureDoorsWalkable` and `ensureGroundConnectivity` filling
cells back in after the carve.

## SOFT notes — reported, never a rejection

Placement shortfalls (best-effort placement is the agreed behaviour — see
CLAUDE.md), zoner blocks a single cell apart, and sparse ground. They are
printed next to the verdict so a batch can still be eyeballed.

Everything is measured in **visual space** (the data grid rotated 90° CCW), so
the numbers describe the room the way the editor draws it and the game renders
it — not the storage layout.

## Calibration

The rules were not written up front — they were derived one review round at a
time, and the method matters more than the current list:

1. **Start from what ships.** The first draft of the rules was run over the 23
   templates already in the project: 13 of them failed, which meant the bar was
   stricter than the game's own content. Anything that flags shipped rooms at a
   similar rate (ground fill, void rectangularity as a *ratio*, quadrant
   balance) is not a quality signal — it is a style the project already uses.
   What survived was what every shipped room passes: zero spurs, zero islands.
2. **Show one room at a time and record the verdict.** Each round generated a
   room, printed a judgement with its reasons, and took a correction. The
   corrections are the rules: zoner must be 2×2, zoner must not be where the
   player walks in, holes must not be 2×1, teaching/building/release want more
   holes, release wants more cover, spacing between zoner blocks is optional.
3. **Put the threshold where the data separates the cases**, not at a round
   number. The void rule is `bbox fill == 1.0` rather than `> 0.7` because the
   accepted rooms were all at exactly 1.0. The density rule is a half-to-half
   ratio at 1.8 because the quadrant ratio could not tell the accepted rooms
   from the rejected ones.
4. **Prefer fixing the generator over filtering its output.** When
   non-rectangular voids failed 80% of rooms, the fix was `rectClearOfVoid` in
   `fullroom.go`, not a laxer rule. Acceptance rules are for what the generator
   cannot guarantee.
5. **"能说出毛病就是不行."** If a defect can be named, the room is rejected —
   there is no passing grade for "acceptable with a flaw". Every rule here
   started as a named defect in a specific room.

Placement shortfalls are the one thing explicitly exempted: the stage counts are
targets, placement is best-effort, and a room that cannot hold its count reports
the gap instead of being thrown away.
