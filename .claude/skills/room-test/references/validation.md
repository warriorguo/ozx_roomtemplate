# Validation Rules for Room Generation Results

All checks are applied to `response.payload`. A result FAILS if any check returns false.
Warnings are non-blocking observations worth noting but not counted as failures.

---

## 1. Structural Validity

### 1.1 Layer Dimensions
Every layer must be exactly `height` rows × `width` columns.
Layers to check: `ground`, `softEdge`, `bridge`, `rail`, `static`, `chaser`, `zoner`, `dps`, `mobAir`, `mainPath`

```
for each layer:
  assert len(layer) == height          # row count
  for each row in layer:
    assert len(row) == width           # column count
```

### 1.2 Cell Values
Every cell in every layer must be 0 or 1.

```
for each layer, row, cell:
  assert cell in {0, 1}
```

### 1.3 Meta Fields
```
assert payload.meta.width == 20
assert payload.meta.height == 12
assert payload.meta.version == 1
```

---

## 2. Ground Connectivity

All cells where `ground[y][x] == 1` must form a **single connected component** (4-connectivity: up/down/left/right only).

```python
def check_ground_connectivity(ground, width, height):
    # Find first ground cell
    start = None
    for y in range(height):
        for x in range(width):
            if ground[y][x] == 1:
                start = (x, y)
                break
        if start: break

    if not start:
        return False  # No ground at all — fail

    # BFS flood fill
    visited = set()
    queue = [start]
    while queue:
        x, y = queue.pop()
        if (x, y) in visited: continue
        visited.add((x, y))
        for dx, dy in [(0,1),(0,-1),(1,0),(-1,0)]:
            nx, ny = x+dx, y+dy
            if 0 <= nx < width and 0 <= ny < height and ground[ny][nx] == 1:
                queue.append((nx, ny))

    # Count total ground cells
    total = sum(ground[y][x] for y in range(height) for x in range(width))
    return len(visited) == total  # All connected
```

**Note**: For bridge rooms, ground connectivity may be intentionally broken (floating islands).
In that case, verify that every disconnected island has at least one adjacent bridge tile.

### 2.1 Door Anchors Are Walkable (ORT-105)

Every **enabled** door's anchor cell must be walkable and reachable from the
other doors. The anchor is the midpoint of that wall:

| Door | Anchor |
|------|--------|
| top | `(width // 2, 0)` |
| bottom | `(width // 2, height - 1)` |
| left | `(0, height // 2)` |
| right | `(width - 1, height // 2)` |

```python
def check_doors_walkable(payload, width, height):
    ground = payload['ground']
    bridge = payload.get('bridge') or [[0] * width for _ in range(height)]
    anchors = {
        'top':    (width // 2, 0),
        'bottom': (width // 2, height - 1),
        'left':   (0, height // 2),
        'right':  (width - 1, height // 2),
    }
    failures = []
    for door, (x, y) in anchors.items():
        if payload['doors'][door] != 1:
            continue
        if ground[y][x] != 1 and bridge[y][x] != 1:
            failures.append(f"door {door} anchor ({x},{y}) is not walkable")
    return failures
```

Combined with §2 (all ground is one component) this is enough: if every anchor
is walkable and all walkable ground is connected, every door reaches every other
door. Generation enforces this via `ensureDoorsWalkable`, and `ComputeMainPath`
fails the request outright if it does not hold (ORT-106) — so a saved template
that violates it came from hand editing, not the generator.

---

## 2a. SoftEdge Anchoring (ORT-116)

For every cell where `softEdge[y][x] == 1`:

1. `ground[y][x]` must be `0` — a soft edge never overlaps ground.
2. The cell must be **anchored** to the ground. Anchoring is a least fixpoint
   over the whole softEdge layer, not a per-cell adjacency test:
   - **base** — the cell is orthogonally adjacent to a `ground == 1` tile
   - **right+down** — the cells to the *visual* right and below are both anchored
   - **left+up** — the cells to the *visual* left and above are both anchored

**The pairs are in visual space; the arrays are in data space** (ORT-117). The
editor renders the grid rotated 90 degrees CCW (ORT-111), so the offsets to walk
in the stored arrays are the transpose of how the rule reads:

| rule | visual | data offsets |
|------|--------|--------------|
| 1 | right + below | `(x, y+1)` and `(x-1, y)` |
| 2 | left + above | `(x, y-1)` and `(x+1, y)` |

Using the literal data-space reading instead inverts the rule into its mirror
image. The rule is deliberately asymmetric — only two of the four corner pairs
count — so a validator that accepts any two adjacent anchored neighbours is
wrong, not more permissive.

The propagation rules borrow only from cells that are themselves anchored, so a
soft edge patch floating in void with no ground contact anywhere stays
unanchored however large it is. Cells with `softEdge == 0` never lend support,
and a diagonal neighbour alone is never enough.

This lets a soft edge patch be thicker than one cell as long as its outer
boundary reaches ground. Generation only ever emits ground-adjacent 1xN / Nx1
strips, so generated rooms only exercise the base rule — the propagation rules
matter for hand-edited rooms.

```python
def compute_softedge_support(ground, soft_edge, width, height):
    """Return a [height][width] grid of booleans: is this soft edge anchored?"""
    supported = [[False] * width for _ in range(height)]
    queue = []

    def adjacent_to_ground(x, y):
        for dx, dy in ((-1, 0), (1, 0), (0, -1), (0, 1)):
            nx, ny = x + dx, y + dy
            if 0 <= nx < width and 0 <= ny < height and ground[ny][nx] == 1:
                return True
        return False

    # Seed with the base rule
    for y in range(height):
        for x in range(width):
            if soft_edge[y][x] == 1 and adjacent_to_ground(x, y):
                supported[y][x] = True
                queue.append((x, y))

    def is_supported(x, y):
        return 0 <= x < width and 0 <= y < height and supported[y][x]

    def can_borrow(x, y):
        # Data-space spelling of "visual right and below" / "visual left and above"
        if soft_edge[y][x] != 1:
            return False
        return ((is_supported(x, y + 1) and is_supported(x - 1, y))
                or (is_supported(x, y - 1) and is_supported(x + 1, y)))

    # Relax until the fixpoint is reached
    while queue:
        cx, cy = queue.pop()
        for nx, ny in ((cx - 1, cy), (cx, cy - 1), (cx + 1, cy), (cx, cy + 1)):
            if not (0 <= nx < width and 0 <= ny < height):
                continue
            if supported[ny][nx] or not can_borrow(nx, ny):
                continue
            supported[ny][nx] = True
            queue.append((nx, ny))

    return supported
```

**Failure messages**: `"soft edge cannot overlap with ground"` ·
`"soft edge has no ground anchor"`

---

## 3. Bridge Validity

For every cell where `bridge[y][x] == 1`, the 2×2 block starting at the top-left of the
bridge tile must have at least one **full edge** (all cells in that edge row/column) touching ground.

Since bridges are always 2×2, find each bridge block's top-left corner first:

```python
def find_bridge_blocks(bridge, width, height):
    """Find all unique 2x2 bridge blocks (top-left corners)."""
    visited = set()
    blocks = []
    for y in range(height):
        for x in range(width):
            if bridge[y][x] == 1 and (x, y) not in visited:
                # Assume this is top-left; verify
                if (x+1 < width and bridge[y][x+1] == 1 and
                    y+1 < height and bridge[y+1][x] == 1 and
                    bridge[y+1][x+1] == 1):
                    blocks.append((x, y))
                    visited.update([(x,y),(x+1,y),(x,y+1),(x+1,y+1)])
    return blocks

def check_bridge_block(bx, by, ground, width, height):
    """
    A bridge block at (bx, by) is valid if AT LEAST ONE of its 4 edges
    has ALL adjacent external cells as ground=1.

    Edge definitions for 2x2 block at (bx, by):
    - Top edge: row by-1, columns bx and bx+1
    - Bottom edge: row by+2, columns bx and bx+1
    - Left edge: col bx-1, rows by and by+1
    - Right edge: col bx+2, rows by and by+1
    """
    edges = {
        'top':    [(bx, by-1), (bx+1, by-1)],
        'bottom': [(bx, by+2), (bx+1, by+2)],
        'left':   [(bx-1, by), (bx-1, by+1)],
        'right':  [(bx+2, by), (bx+2, by+1)],
    }

    for edge_name, cells in edges.items():
        # All cells of this edge must be in bounds AND ground=1
        if all(
            0 <= x < width and 0 <= y < height and ground[y][x] == 1
            for x, y in cells
        ):
            return True, edge_name  # Valid — has a full ground edge

    return False, None  # Floating — no full edge touches ground
```

**Failure message**: `"Bridge block at ({bx},{by}) is floating — no full edge touches ground"`

---

## 4. Entity Layer Constraints

### 4.1 Must be on Ground
`static`, `chaser`, `zoner`, `dps` cells must only exist where `ground == 1`.

```
for each layer in [static, chaser, zoner, dps]:
  for each (x, y) where layer[y][x] == 1:
    assert ground[y][x] == 1
```

`mobAir` has no ground requirement — skip this check for mobAir.

### 4.2 No Invalid Overlaps
Per validation rules. Chaser, zoner and dps all collide with static (ORT-119) —
dps used to be exempt in the frontend rule and none of the three were checked by
the backend at all:
- `static==1`: cannot overlap chaser, zoner, dps, bridge, rail
- `chaser==1`: cannot overlap static, bridge, rail, zoner
- `zoner==1`: cannot overlap static, bridge, rail, chaser
- `dps==1`: cannot overlap static, bridge, rail, zoner (CAN coexist with chaser)
- `mobAir==1`: **no overlap rule at all** (ORT-121). Neither validator checks
  mobAir, and hand-authored rooms stack it on zoner/chaser/dps/static freely.
  Generation still keeps it clear as a dispersion preference — assert that only
  against `/generate` output (§5a-4), never against a loaded template.

```python
overlap_rules = {
    'static':  ['chaser', 'zoner', 'dps', 'bridge', 'rail'],
    'chaser':  ['static', 'bridge', 'rail', 'zoner'],
    'zoner':   ['static', 'bridge', 'rail', 'chaser'],
    'dps':     ['static', 'bridge', 'rail', 'zoner'],
    # mobAir omitted deliberately (ORT-121): it has no overlap rule.
}

for layer_name, forbidden in overlap_rules.items():
    layer = payload[layer_name]
    for f in forbidden:
        other = payload[f]
        for y in range(height):
            for x in range(width):
                if layer[y][x] == 1 and other[y][x] == 1:
                    FAIL: f"{layer_name} overlaps {f} at ({x},{y})"
```

---

## 5. Stage Rules (Enemy Count Ranges)

For each `stageType`, validate that actual entity counts match expected ranges.

Count entities by summing all 1s in each layer — **except `zoner`**, which is
counted in 2×2 blocks. See §5c.

```
stageRanges = {
    "teaching":  { dps: (4,6),   chaser: (2,2),   zoner: (1,1),  mobAir: (6,6)  },
    "building":  { dps: (4,6),   chaser: (4,6),   zoner: (1,1),  mobAir: (6,6)  },
    "pressure":  { dps: (8,12),  chaser: (8,10),  zoner: (2,2),  mobAir: (6,12) },
    "peak":      { dps: (8,12),  chaser: (8,10),  zoner: (2,3),  mobAir: (12,18)},
    "release":   { dps: (2,4),   chaser: (2,2),   zoner: (1,1),  mobAir: (6,6)  },
    "boss":      { dps: (0,0),  chaser: (0,0),  zoner: (0,0),  mobAir: (0,0) },
}
```

If `stageType` is empty or not provided, skip this check.

### 5a. Stage Static Counts (ORT-100)

When a `stageType` is supplied, the stage also drives `staticCount` — the
request's `staticCount` is **overridden**, not merged. The high-pressure stages
get roughly one third the cover of the low-pressure ones:

```
stageStaticBlocks = {
    "start":    (0,0),   # no enemies, no cover needed
    "teaching": (6,9),
    "building": (6,9),
    "pressure": (2,3),   # ~1/3 of the baseline
    "peak":     (2,3),   # ~1/3 of the baseline
    "release":  (6,9),
    "boss":     (0,0),   # 6x6 clear center arena
}
```

These are counts of **2×2 blocks**, not cells. Blocks never touch, so a
successful placement of N blocks yields `4 * N` cells in the `static` layer.

**Do not assert an exact cell count.** Placement is best-effort: door forbidden
zones can exhaust the valid 2×2 sites and leave the layer short of target (open
bug ORT-40). Treat the range as an upper bound:

```
static_cells = sum(payload['static'][y][x] for y in range(h) for x in range(w))
lo, hi = stageStaticBlocks[st]
assert static_cells % 4 == 0, "static cells not whole 2x2 blocks"
assert static_cells <= hi * 4, f"stage {st}: {static_cells//4} blocks, max {hi}"
# under-placement is ORT-40, report it but do not fail the run
if static_cells < lo * 4:
    warnings.append(f"stage {st}: only {static_cells//4} blocks, expected >= {lo} (ORT-40)")
```

With an empty `stageType` the request's `staticCount` is used verbatim.

### 5a-2. Stage Static Placement Strategy (ORT-99)

The stage also selects *where* the blocks go, via `StagePlacementHints.StaticDisperse`:

| Stage | Strategy |
|---|---|
| start, teaching, building, release | alternating centre-outward / edge-inward scatter (the long-standing default) |
| pressure, peak | edge-first seeding, then farthest-point dispersion |

pressure/peak carry the heaviest enemy loads, so their cover is pushed to the
perimeter to leave the middle open. Expect their statics to show a **lower mean
distance from the nearest wall** and a **larger mean nearest-neighbour spacing**
than the default path. This is a distributional property — do not assert it on a
single generated room; average over many trials if you check it at all.

All the hard constraints are unchanged by strategy: 8-directional non-contact,
door forbidden diamond, no overlap with softEdge/bridge/rail, and door-to-door
connectivity after placement.

### 5a-3. MobAir Distribution (ORT-104)

MobAir is placed **centre-seeded and evenly distributed**, deterministically —
the same room always produces the same layer, so re-generating an identical
request and getting a different `mobAir` is a regression.

Two properties are worth checking:

1. **The centre is occupied.** The valid cell closest to the room centre always
   carries a mobAir. "Valid" here means the cell passes the §5a-4 constraints
   below in an otherwise-empty mobAir layer.
2. **Placements are dispersed, not clustered.** Measure with the Clark-Evans
   nearest-neighbour index — observed mean nearest-neighbour distance divided by
   the mean expected from scattering the same number of points at random over
   the same placeable area:

   ```python
   expected = 0.5 * math.sqrt(placeable_area / len(points))
   R = observed_mean_nn / expected     # <1 clustered, 1 random, >1 dispersed
   ```

   Expect **R > 1.25**; measured values run 1.4–1.9. Do **not** substitute a
   plain variance-of-spacing check — a tight cluster has uniformly small
   spacings and scores well on variance alone. The density-driven placement this
   replaced measured R = 0.95 (indistinguishable from random) while scoring
   *better* on raw variance.

   Average over many trials; a single room is not a distribution.

The old zoner/chaser "dense area" preference still exists, but only as a
tiebreak between the cells one grid slot could snap to. It no longer moves mobs
across the room, so do not expect mobAir to track enemy clusters.

### 5a-4. MobAir Hard Constraints

Unchanged by ORT-104, and worth asserting on every room:

- door distance ≥ 4 (Manhattan), edge distance ≥ 2
- no 8-directional adjacency between mobAir cells
- no overlap with `chaser`, `zoner`, `dps`, `static` — mobAir runs last in the
  pipeline, so any overlap means a later layer was placed on top of it, not the
  reverse. This was a real defect in fullroom's grouped placement before ORT-104
  (mobAir ran inside the per-group loop); treat any overlap as a regression.

  **Generation invariant only, not a validation rule** (ORT-121). Assert it
  against `/generate` output; never against a loaded template.

### 5b. Room Size Is Unconstrained (ORT-109)

No stage constrains room dimensions. Any room size may be paired with any stage,
and a generate request is never rejected for being too small.

ORT-102 previously gave `pressure` an 18×10 minimum and `peak` a 20×12 one.
ORT-108 cut those two stages' counts back instead (pressure chaser 12-16 → 8-10;
peak dps 12-24 → 8-12, chaser 12-16 → 8-10, zoner 4-6 → 2-3, mobAir 18 → 12-18),
and the limits were removed.

**Testing this**: a `pressure` or `peak` generate request at 16×8 must return
200 with a template. Treat a 400 mentioning `room size` as a regression.

**Small rooms are permitted, not guaranteed clean.** Measured at ORT-108's
counts over 200 rooms per size, the share containing an adjacent same-category
spawn pair (the ORT-93 crash condition) is:

| Room | pressure | peak |
|------|----------|------|
| 16×8 | 6.0% | 16.5% |
| 16×10 | 0.5% | 4.5% |
| 18×10 | 0.0% | 2.5% |
| 20×12 | 0.0% | 0.5% |
| 24×14 | 0.0% | 0.0% |

So §4.3 (spacing) is worth checking especially closely on small rooms, and a
violation there is an ORT-93 finding rather than a size-limit one.

Small rooms also under-place: the generate response carries a `warnings` array
listing every layer that fell short (ORT-110). Treat it as **expected
information, not a failure** — but a count outside its stage range with **no**
corresponding warning *is* a failure, because it means the shortfall went
unreported.

```python
def check_shortfall_reporting(response, payload, stage_ranges, stage):
    """A count below its stage minimum must be explained by a warning."""
    warned = {w['layer'] for w in response.get('warnings') or []}
    failures = []
    for layer, (lo, _hi) in stage_ranges.get(stage, {}).items():
        placed = count_zoner_units(payload) if layer == 'zoner' else count_cells(payload[layer])
        if placed < lo and layer not in warned:
            failures.append(f"{layer} placed {placed} (stage min {lo}) with no warning")
    return failures
```

**Note**: Counts here are number of *spawners*, not enemy sprites. `chaser`,
`dps` and `mobAir` occupy one cell per spawner; `zoner` occupies a 2×2 block per
spawner (see §5c).

---

### 5c. Zoner Footprint (ORT-103)

A zoner occupies a **2×2 block**, falling back to **1×1** only where no 2×2 site
fits. The game collapses a connected block into a single spawn, so one zoner is
four cells — summing 1s in the `zoner` layer over-counts by 4×.

Count zoner spawners as **8-connected groups**, and validate the shape of each:

```python
def zoner_groups(layer, width, height):
    seen, groups = set(), []
    for y in range(height):
        for x in range(width):
            if layer[y][x] != 1 or (x, y) in seen:
                continue
            group, queue = [(x, y)], [(x, y)]
            seen.add((x, y))
            while queue:
                cx, cy = queue.pop()
                for dy in (-1, 0, 1):
                    for dx in (-1, 0, 1):
                        nx, ny = cx + dx, cy + dy
                        if 0 <= nx < width and 0 <= ny < height \
                                and layer[ny][nx] == 1 and (nx, ny) not in seen:
                            seen.add((nx, ny))
                            group.append((nx, ny))
                            queue.append((nx, ny))
            groups.append(group)
    return groups

groups = zoner_groups(payload['zoner'], width, height)

# Count: this is the number the stage range in §5 applies to.
zoner_count = len(groups)

# Shape: every group is a solid 2x2 block or a lone fallback cell. Anything
# else means two zoners are touching — the ORT-93 spacing violation, which
# shows up here because touching blocks merge into one group.
for g in groups:
    xs, ys = [p[0] for p in g], [p[1] for p in g]
    is_1x1 = len(g) == 1
    is_2x2 = len(g) == 4 and max(xs) - min(xs) == 1 and max(ys) - min(ys) == 1
    if not (is_1x1 or is_2x2):
        FAIL: f"zoner group of {len(g)} cells at {min(xs)},{min(ys)} is neither 2x2 nor 1x1"
```

Every cell of a block must independently satisfy the per-cell zoner rules in §3
and §4 (ground, overlap, door forbidden zone) — check them cell-by-cell as usual.

1×1 fallbacks are rare in practice (measured under 1% of spawns at every stage
minimum); a room where most zoners are 1×1 signals the 2×2 search is failing and
is worth investigating even though it is technically legal.

---

## 6. Door Forbidden Zone

No entity cells (`chaser`, `zoner`, `dps`, `static`, `mobAir`) may appear within
Manhattan distance 2 of any door center.

Door centers for a 20×12 grid:
```
top:    (10, 0)
bottom: (10, 11)
left:   (0, 6)
right:  (19, 6)
```

Only check doors that are enabled (where `payload.doors.{direction} == 1`).
`payload.doors` already reflects any explicit `payload.doorOverrides` (a side
present there forces open/closed; absent sides fall back to ground
connectivity), so validate against `payload.doors`, not the raw overrides.

```python
def manhattan(x1, y1, x2, y2):
    return abs(x1-x2) + abs(y1-y2)

door_centers = {
    'top': (10, 0), 'bottom': (10, 11),
    'left': (0, 6), 'right': (19, 6)
}

entity_layers = ['chaser', 'zoner', 'dps', 'static', 'mobAir']

for door, (dx, dy) in door_centers.items():
    if payload['doors'][door] != 1:
        continue
    for layer_name in entity_layers:
        layer = payload[layer_name]
        for y in range(height):
            for x in range(width):
                if layer[y][x] == 1 and manhattan(x, y, dx, dy) <= 2:
                    FAIL: f"{layer_name} at ({x},{y}) within radius 2 of {door} door"
```

---

## 7. Boss Room Center Clear Zone

When `stageType == "boss"`, the center 6×6 area must have `static == 0` for all cells.

Center 6×6 for 20×12 grid:
- x: 7 to 12 (inclusive)
- y: 3 to 8 (inclusive)

```python
if stageType == "boss":
    for y in range(3, 9):
        for x in range(7, 13):
            assert static[y][x] == 0, f"static at ({x},{y}) in boss center zone"
```

---

## Warning Conditions (non-blocking)

Record these as warnings, not failures:

- **Low bridge count**: bridge room has 0 bridge tiles (no error, but suspicious)
- **Zero enemies with non-release stage**: unexpected but not invalid
- **All entities clustered**: most entities within 3 cells of each other (poor distribution)
- **softEdge covers door**: softEdge cell adjacent to door position (may block path)

---

## Quick Validation Script (Python)

For batch agents, here's a compact validator to embed in the agent prompt:

```python
import json

def zoner_groups(layer, width, height):
    """8-connected groups of zoner cells. One group = one zoner spawn (ORT-103)."""
    seen, groups = set(), []
    for y in range(height):
        for x in range(width):
            if layer[y][x] != 1 or (x, y) in seen:
                continue
            group, queue = [(x, y)], [(x, y)]
            seen.add((x, y))
            while queue:
                cx, cy = queue.pop()
                for dy in (-1, 0, 1):
                    for dx in (-1, 0, 1):
                        nx, ny = cx + dx, cy + dy
                        if 0 <= nx < width and 0 <= ny < height \
                                and layer[ny][nx] == 1 and (nx, ny) not in seen:
                            seen.add((nx, ny))
                            group.append((nx, ny))
                            queue.append((nx, ny))
            groups.append(group)
    return groups

def validate(payload, expected_stage=None):
    w, h = payload['meta']['width'], payload['meta']['height']
    layers = ['ground','softEdge','bridge','rail','static','chaser','zoner','dps','mobAir','mainPath']
    failures = []
    warnings = []

    # 1. Dimensions + cell values
    for name in layers:
        layer = payload[name]
        if len(layer) != h:
            failures.append(f"{name}: expected {h} rows, got {len(layer)}")
            continue
        for y, row in enumerate(layer):
            if len(row) != w:
                failures.append(f"{name}[{y}]: expected {w} cols, got {len(row)}")
            for x, v in enumerate(row):
                if v not in (0, 1):
                    failures.append(f"{name}[{y}][{x}] = {v}, expected 0 or 1")

    ground = payload['ground']
    bridge = payload['bridge']

    # 2. Ground connectivity (BFS)
    starts = [(x,y) for y in range(h) for x in range(w) if ground[y][x]==1]
    if starts:
        visited, queue = set(), [starts[0]]
        while queue:
            x,y = queue.pop()
            if (x,y) in visited: continue
            visited.add((x,y))
            for dx,dy in [(0,1),(0,-1),(1,0),(-1,0)]:
                nx,ny=x+dx,y+dy
                if 0<=nx<w and 0<=ny<h and ground[ny][nx]==1:
                    queue.append((nx,ny))
        if len(visited) != len(starts):
            failures.append(f"Ground not connected: {len(visited)} reachable of {len(starts)} total")

    # 3. Bridge validity
    seen = set()
    for y in range(h-1):
        for x in range(w-1):
            if bridge[y][x]==1 and (x,y) not in seen:
                if (x+1<w and bridge[y][x+1]==1 and bridge[y+1][x]==1 and bridge[y+1][x+1]==1):
                    seen.update([(x,y),(x+1,y),(x,y+1),(x+1,y+1)])
                    bx,by=x,y
                    edges = {
                        'top':    [(bx,by-1),(bx+1,by-1)],
                        'bottom': [(bx,by+2),(bx+1,by+2)],
                        'left':   [(bx-1,by),(bx-1,by+1)],
                        'right':  [(bx+2,by),(bx+2,by+1)],
                    }
                    valid = any(
                        all(0<=cx<w and 0<=cy<h and ground[cy][cx]==1 for cx,cy in pts)
                        for pts in edges.values()
                    )
                    if not valid:
                        failures.append(f"Bridge block at ({bx},{by}) floating — no full edge on ground")

    # 4. Entity on ground + overlaps
    ol_rules = {
        'static':['chaser','zoner','dps','bridge','rail'],
        'chaser':['static','bridge','rail','zoner'],
        'zoner':['static','bridge','rail','chaser'],
        'dps':['bridge','rail','zoner'],
        'mobAir':['chaser','zoner','dps','static'],
    }
    for lname, forbidden in ol_rules.items():
        layer = payload[lname]
        if lname != 'mobAir':
            for y in range(h):
                for x in range(w):
                    if layer[y][x]==1 and ground[y][x]!=1:
                        failures.append(f"{lname}[{y}][{x}]=1 but ground=0")
        for f in forbidden:
            other = payload[f]
            for y in range(h):
                for x in range(w):
                    if layer[y][x]==1 and other[y][x]==1:
                        failures.append(f"{lname} overlaps {f} at ({x},{y})")

    # 5. Stage rules
    stage_ranges = {
        'teaching': {'dps':(4,6),'chaser':(2,2),'zoner':(1,1),'mobAir':(6,6)},
        'building': {'dps':(4,6),'chaser':(4,6),'zoner':(1,1),'mobAir':(6,6)},
        'pressure': {'dps':(8,12),'chaser':(8,10),'zoner':(2,2),'mobAir':(6,12)},
        'peak':     {'dps':(8,12),'chaser':(8,10),'zoner':(2,3),'mobAir':(12,18)},
        'release':  {'dps':(2,4),'chaser':(2,2),'zoner':(1,1),'mobAir':(6,6)},
        'boss':     {'dps':(0,0),'chaser':(0,0),'zoner':(0,0),'mobAir':(0,0)},
    }
    st = expected_stage or (payload.get('stageType') or '')
    if st and st in stage_ranges:
        for ename, (lo, hi) in stage_ranges[st].items():
            if ename == 'zoner':
                # A zoner is a 2x2 block collapsed into one spawn (ORT-103),
                # so count 8-connected groups rather than cells.
                count = len(zoner_groups(payload['zoner'], w, h))
            else:
                count = sum(payload[ename][y][x] for y in range(h) for x in range(w))
            if not (lo <= count <= hi):
                failures.append(f"Stage {st}: {ename} count={count}, expected [{lo},{hi}]")

    # 5c. Zoner footprint: every group is a solid 2x2 block or a lone fallback
    # cell. Any other shape means two zoners touch (ORT-93 spacing violation).
    for g in zoner_groups(payload['zoner'], w, h):
        xs, ys = [p[0] for p in g], [p[1] for p in g]
        is_1x1 = len(g) == 1
        is_2x2 = len(g) == 4 and max(xs)-min(xs) == 1 and max(ys)-min(ys) == 1
        if not (is_1x1 or is_2x2):
            failures.append(
                f"zoner group of {len(g)} cells at ({min(xs)},{min(ys)}) is neither 2x2 nor 1x1")

    # 6. Door forbidden zone
    door_pos = {'top':(10,0),'bottom':(10,11),'left':(0,6),'right':(19,6)}
    doors = payload.get('doors', {})
    entity_layers = ['chaser','zoner','dps','static','mobAir']
    for dname, (dx,dy) in door_pos.items():
        if doors.get(dname, 0) != 1:
            continue
        for lname in entity_layers:
            layer = payload[lname]
            for y in range(h):
                for x in range(w):
                    if layer[y][x]==1 and abs(x-dx)+abs(y-dy) <= 2:
                        failures.append(f"{lname} at ({x},{y}) within radius 2 of {dname} door")

    # 7. Boss center clear
    if st == 'boss':
        for y in range(3, 9):
            for x in range(7, 13):
                if payload['static'][y][x] == 1:
                    failures.append(f"static at ({x},{y}) in boss 6x6 center zone")

    # Warnings
    bridge_count = sum(bridge[y][x] for y in range(h) for x in range(w))
    if payload.get('roomShape') == 'bridge' and bridge_count == 0:
        warnings.append("Bridge room has 0 bridge tiles")

    return failures, warnings
```
