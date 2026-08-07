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
Per validation rules:
- `static==1`: cannot overlap chaser, zoner, dps, bridge, rail
- `chaser==1`: cannot overlap static, bridge, rail, zoner
- `zoner==1`: cannot overlap static, bridge, rail, chaser
- `dps==1`: cannot overlap bridge, rail, zoner (CAN coexist with chaser, static)
- `mobAir==1`: cannot overlap chaser, zoner, dps, static

```python
overlap_rules = {
    'static':  ['chaser', 'zoner', 'dps', 'bridge', 'rail'],
    'chaser':  ['static', 'bridge', 'rail', 'zoner'],
    'zoner':   ['static', 'bridge', 'rail', 'chaser'],
    'dps':     ['bridge', 'rail', 'zoner'],
    'mobAir':  ['chaser', 'zoner', 'dps', 'static'],
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

Count entities by summing all 1s in each layer.

```
stageRanges = {
    "teaching":  { dps: (4,6),   chaser: (2,2),   zoner: (1,1),  mobAir: (6,6)  },
    "building":  { dps: (4,6),   chaser: (4,6),   zoner: (1,1),  mobAir: (6,6)  },
    "pressure":  { dps: (8,12),  chaser: (12,16), zoner: (2,2),  mobAir: (6,12) },
    "peak":      { dps: (12,24), chaser: (12,16), zoner: (4,6),  mobAir: (18,18)},
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

### 5b. Stage Minimum Room Size

Some stages require a minimum room size; the generator rejects smaller rooms
before building any layer, so a generate request that violates this returns an
error rather than a template.

```
stageMinSize = {
    "pressure": (18, 10),   # width, height
    "peak":     (20, 12),
}
```

All other stages are unconstrained.

Rationale: below these dimensions the strict placement pass cannot satisfy the
stage's enemy minimums, and the relaxed fallback meets them by dropping the
8-directional spacing constraint — producing adjacent same-category spawn tiles
that crash the game for Zoner (ORT-93). See ORT-102.

**Testing this**: a `pressure` or `peak` generate request below the minimum must
fail with HTTP 400 and a message matching
`stage <name> room size: requires a room of at least WxH, got WxH`.
Treat a 200 response for an undersized room as a regression.

**Note**: Counts here are number of *spawner cells*, not enemy sprites. Each mob
occupies one cell in its layer.

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
        'pressure': {'dps':(8,12),'chaser':(12,16),'zoner':(2,2),'mobAir':(6,12)},
        'peak':     {'dps':(12,24),'chaser':(12,16),'zoner':(4,6),'mobAir':(18,18)},
        'release':  {'dps':(2,4),'chaser':(2,2),'zoner':(1,1),'mobAir':(6,6)},
        'boss':     {'dps':(0,0),'chaser':(0,0),'zoner':(0,0),'mobAir':(0,0)},
    }
    st = expected_stage or (payload.get('stageType') or '')
    if st and st in stage_ranges:
        for ename, (lo, hi) in stage_ranges[st].items():
            count = sum(payload[ename][y][x] for y in range(h) for x in range(w))
            if not (lo <= count <= hi):
                failures.append(f"Stage {st}: {ename} count={count}, expected [{lo},{hi}]")

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
