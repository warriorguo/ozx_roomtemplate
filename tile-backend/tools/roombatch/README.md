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

| rule | why |
|------|-----|
| every zoner spawn is a full 2×2 block | explicit instruction; the 1×1 fallback was removed from the generator |
| no zoner in the room's visual bottom third | the player walks in from there and fires automatically |
| no 1-cell ground spurs | all 23 hand-authored rooms have zero |
| no unreachable ground island | a slab the player cannot get to, sometimes carrying spawns |

Hard-rule pass rate at 16×10, 25 rooms per stage: teaching 18/25, building
22/25, pressure 22/25, peak 20/25, release 20/25. Every failure observed was a
ground spur, so ~1.3 attempts per accepted room.

## SOFT notes — reported, never a rejection

Placement shortfalls (best-effort placement is the agreed behaviour — see
CLAUDE.md), quadrant balance, ragged carved voids, sparse ground, static
crowding. They are printed next to the verdict so a batch can be eyeballed.

Everything is measured in **visual space** (the data grid rotated 90° CCW), so
the numbers describe the room the way the editor draws it and the game renders
it — not the storage layout.

## Calibration

The soft rules were originally hard rules. They were demoted after running the
same checks over the 23 templates already in the project: 13 of them failed, so
the bar was stricter than what ships. The four hard rules are the ones every
shipped room passes, plus the two zoner rules given in review.
