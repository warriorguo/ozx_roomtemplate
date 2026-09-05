---
name: room-generator
description: Generate game room templates (full/bridge/platform) via backend API. Use this skill when the user wants to generate a room, create a room template, test room generation, or batch-generate rooms. Triggers on phrases like "generate a room", "create room template", "test fullroom generation", "generate bridge room", "make a platform room", or any room generation task.
---

# Room Generator

Generate tile-based game room templates by calling the backend generation API. Produces a multi-layer room with ground, softEdge, bridge, rail, static, chaser, zoner, dps, and mobAir layers.

## Parameters

All parameters have sensible defaults. Show them to the user and let them modify before generating.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `endpoint` | string | `https://ozx-roomtpl.local.playquota.com/api/v1` | Backend API base URL |
| `roomType` | string | `"full"` | Room type: `"full"` (ground almost fully filled), `"bridge"` (narrow corridor paths), or `"platform"` (large platform blocks) |
| `width` | int | `20` | Room width in tiles (4-200) |
| `height` | int | `12` | Room height in tiles (4-200) |
| `doors` | string[] | `["top","right","bottom","left"]` | Doors to connect. At least 2 required. Options: `top`, `right`, `bottom`, `left` |
| `stageType` | string | `""` | Stage type: `"start"`, `"teaching"`, `"building"`, `"pressure"`, `"peak"`, `"release"`, `"boss"`, or empty (defaults to `"default"` in output). Controls enemy count ranges **and `staticCount`** — when set, it overrides the counts below. |
| `roomCategory` | string | `"normal"` | Room category: `"normal"`, `"basement"`, `"test"`, `"cave"`. Passed through to output. |
| `softEdgeCount` | int | `3` | Number of soft edge strips to place in void notches |
| `railEnabled` | bool | `true` | Whether to generate a rail loop on the ground |
| `staticCount` | int | `8` | Number of 2x2 static obstacle blocks. **Ignored when `stageType` is set** — the stage supplies it: teaching/building/release 6–9, pressure/peak 2–3, start/boss 0. |
| `chaserCount` | int | `4` | Number of chaser placements (melee enemies near main path) |
| `zonerCount` | int | `2` | Number of zoner placements (area control enemies). Counted in **spawns, not cells** — each zoner occupies a 2x2 block, so `zonerCount: 2` produces 8 cells. |
| `dpsCount` | int | `4` | Number of DPS placements (ranged damage enemies) |
| `mobAirCount` | int | `10` | Number of air mob spawn points |
| `outputPath` | string | *(ask user)* | File path to save the full JSON response. If not provided, ask the user where to save it. |

## Workflow

### Step 1: Show parameters and confirm

Display the current parameter values in a table. Ask the user if they want to change anything before generating. Example:

```
Room Generation Parameters:
  endpoint:       https://ozx-roomtpl.local.playquota.com/api/v1
  roomType:       full
  width:          20
  height:         12
  doors:          [top, right, bottom, left]
  stageType:      (none)
  roomCategory:   normal
  softEdgeCount:  3
  railEnabled:    true
  staticCount:    8
  chaserCount:    4
  zonerCount:     2
  dpsCount:       4
  mobAirCount:    10
  outputPath:     (not set - will ask after generation)
```

If the user provides parameters inline (e.g., "generate a 30x20 bridge room with 2 doors"), parse them and apply.

### Step 2: Call the API

Map `roomType` to API endpoint:
- `"full"` → `POST {endpoint}/generate/fullroom`
- `"bridge"` → `POST {endpoint}/generate/bridge`
- `"platform"` → `POST {endpoint}/generate/platform`

Request body:
```json
{
  "width": <width>,
  "height": <height>,
  "doors": <doors>,
  "stageType": <stageType>,
  "roomCategory": <roomCategory>,
  "softEdgeCount": <softEdgeCount>,
  "railEnabled": <railEnabled>,
  "staticCount": <staticCount>,
  "chaserCount": <chaserCount>,
  "zonerCount": <zonerCount>,
  "dpsCount": <dpsCount>,
  "mobAirCount": <mobAirCount>
}
```

Use `curl` via Bash to call the API. Parse the JSON response with `python3 -c`.

### Step 3: Display ASCII visualization

Render a composite view of the room using these symbols:

| Symbol | Layer | Priority (highest first) |
|--------|-------|--------------------------|
| `R` | rail | 1 |
| `S` | static | 2 |
| `C` | chaser | 3 |
| `Z` | zoner | 4 |
| `D` | dps | 5 |
| `A` | mobAir | 6 |
| `E` | softEdge | 7 |
| `B` | bridge | 8 |
| `█` | ground (walkable) | 9 |
| `·` | void | 10 |

Use this python snippet to render:
```python
import json, sys
data = json.load(sys.stdin)
p = data['payload']
h, w = len(p['ground']), len(p['ground'][0])
shape = p.get('roomShape', 'unknown')
category = p.get('roomCategory', 'unknown')
stage = p.get('stageType', 'none')
print(f"Room: {w}x{h} shape={shape} category={category} stage={stage}")
print()
for y in range(h):
    line = ''
    for x in range(w):
        if p.get('rail') and p['rail'][y][x]: line += 'R'
        elif p['static'][y][x]: line += 'S'
        elif p.get('chaser') and p['chaser'][y][x]: line += 'C'
        elif p.get('zoner') and p['zoner'][y][x]: line += 'Z'
        elif p.get('dps') and p['dps'][y][x]: line += 'D'
        elif p['mobAir'][y][x]: line += 'A'
        elif p.get('softEdge') and p['softEdge'][y][x]: line += 'E'
        elif p.get('bridge') and p['bridge'][y][x]: line += 'B'
        elif p['ground'][y][x]: line += '█'
        else: line += '·'
    print(line)
```

### Step 4: Show debug summary

Extract and display key debug info from the response:

**For full rooms:**
- Corner erase: skipped or combo used + brush size
- Center pits: skipped or pit count + symmetry

**For all room types:**
- Rail: platforms found, loops placed, perimeter
- Static: target vs placed count
- Chaser: target vs placed count
- Zoner: target vs placed count
- DPS: target vs placed count
- MobAir: target vs placed count
- **Warnings**: if `warnings` is present, list each shortfall (`layer: placed N of M`)
  and say the room is under-populated for its stage. Do not treat a 200 response
  as proof the stage's counts were met — see "Placement Shortfalls" below.

### Step 5: Save the result

If the user provides an `outputPath`, save there directly. Otherwise, follow the subfolder convention:

**Subfolder convention** (for OZX Unity project):
When saving to `Assets/StreamingAssets/TilemapData/`, organize by `roomCategory`:
```
TilemapData/
├── normal/      ← roomCategory == "normal" (default)
├── basement/    ← roomCategory == "basement"
├── test/        ← roomCategory == "test"
└── cave/        ← roomCategory == "cave"
```

**Auto-naming**: If the user doesn't specify a filename, generate one from the response:
```
{roomShape}_{stageType}_{openDoors}_{seq}.json
```
- `roomShape`: from payload (`"all"`, `"bridge"`, `"platform"`), or `"none"` if null
- `stageType`: from payload (`"teaching"`, `"building"`, etc.) — always present, defaults to `"default"` when not specified
- `openDoors`: bitmask from payload (Top=1, Right=2, Bottom=4, Left=8). e.g. top+bottom = 5, all doors = 15
- `seq`: two-digit sequence number, auto-incremented by scanning existing files with the same `{roomShape}_{stageType}_{openDoors}_` prefix in the target folder

Examples: `bridge_teaching_5_01.json`, `all_default_15_02.json`, `platform_boss_5_01.json`

**Auto-increment logic**: Use `Glob` to find `{targetDir}/{roomShape}_{stageType}_{openDoors}_*.json`, extract the highest sequence number, and increment by 1. Start at `01` if none exist.

**Before saving**, remap door directions for OZX Unity coordinate convention (transpose + Y-flip):

```python
# Combined: top→left, bottom→right, left→bottom, right→top
DOOR_REMAP = {'top': 'left', 'right': 'top', 'bottom': 'right', 'left': 'bottom'}
BITMASK_REMAP = {1: 8, 2: 1, 4: 2, 8: 4}  # Top=1→Left=8, Right=2→Top=1, Bottom=4→Right=2, Left=8→Bottom=4

# Remap doors object
if 'doors' in payload:
    old = payload['doors']
    payload['doors'] = {DOOR_REMAP[k]: v for k, v in old.items() if k in DOOR_REMAP}

# Remap openDoors bitmask
if 'openDoors' in payload:
    old_mask = payload['openDoors']
    new_mask = 0
    for bit, remapped_bit in BITMASK_REMAP.items():
        if old_mask & bit:
            new_mask |= remapped_bit
    payload['openDoors'] = new_mask
```

Save the **payload only** (not debugInfo) to the file — this is what the game client loads.
Ask the user to confirm the path before writing.

Confirm: "Saved to {outputPath}"

## Response Structure

The API returns this JSON structure:

```json
{
  "payload": {
    "ground": [[0,1,...], ...],      // 2D grid, 0=void, 1=walkable
    "softEdge": [[0,1,...], ...],    // Fills void notches adjacent to ground
    "bridge": [[0,1,...], ...],      // Connects floating islands
    "rail": [[0,1,...], ...],        // Closed loop track on ground
    "static": [[0,1,...], ...],      // 2x2 obstacle blocks
    "chaser": [[0,1,...], ...],      // Melee enemy positions (near main path)
    "zoner": [[0,1,...], ...],       // Area control enemy positions
    "dps": [[0,1,...], ...],         // Ranged damage enemy positions
    "mobAir": [[0,1,...], ...],      // Air mob spawn points (no ground required)
    "mainPath": [[0,1,...], ...],    // Main path through room center
    "doors": {
      "top": 0|1,
      "right": 0|1,
      "bottom": 0|1,
      "left": 0|1
    },
    "doorOverrides": {          // optional; user's explicit open-door whitelist
      "top": 1,                 // side present & =1 => explicitly open; omit => auto
      "left": 1                 // when any side is set, open doors = exactly these
    },                          // (others closed); empty/absent => ground connectivity
    "roomShape": "all"|"bridge"|"platform",
    "roomCategory": "normal"|"basement"|"test"|"cave",
    "stageType": "default"|"start"|"teaching"|"building"|"pressure"|"peak"|"release"|"boss",
    "meta": {
      "name": "full-20x12",
      "version": 1,
      "width": 20,
      "height": 12
    }
  },
  "debugInfo": {
    "ground": { ... },        // Room-type specific ground debug
    "rail": {                 // Rail generation debug
      "skipped": false,
      "platformsFound": 1,
      "railLoops": [{ "platform": "...", "boundingBox": "...", "perimeter": 36 }]
    },
    "softEdge": { "skipped": false, "targetCount": 3, "placedCount": 3, ... },
    "bridgeLayer": { ... },
    "static": { "skipped": false, "targetCount": 8, "placedCount": 8, ... },
    "chaser": { "skipped": false, "targetCount": 4, "placedCount": 4, ... },
    "zoner": { "skipped": false, "targetCount": 2, "placedCount": 2, ... },
    "dps": { "skipped": false, "targetCount": 4, "placedCount": 4, ... },
    "mobAir": { "skipped": false, "targetCount": 10, "placedCount": 10, ... }
  },
  "warnings": [                 // optional; absent when every layer met its target
    {
      "layer": "mobAir",
      "requested": 14,
      "placed": 7,
      "message": "mobAir: placed 7 of 14 spawns — the room ran out of legal sites under the spacing constraint"
    }
  ]
}
```

### Placement Shortfalls (`warnings`)

Placement is best-effort. The strict pass keeps the 8-directional spacing
constraint (ORT-93) and stops when a room runs out of legal sites, and mobAir
has no relaxed fallback at all — so a room smaller than its stage wants comes
back lighter than requested. `warnings` reports each such layer with requested
vs placed; it is **absent entirely** when everything met its target.

Counts are spawns, not cells — a zoner is a 2×2 block (ORT-103).

This matters most since ORT-109 removed the per-stage minimum room sizes: a
16×8 peak room is now generated rather than refused, and `warnings` is how the
caller learns it is under-populated. **Report warnings to the user when
present** — a 200 response is not by itself evidence the room got what the
stage asked for. The room is still valid and saveable.

### Layer Rules

- **ground**: Foundation layer. Full rooms start 100% filled then carve corners/pits. Bridge rooms connect doors with paths. Platform rooms use large rectangular blocks.
- **softEdge**: Placed in void cells adjacent to ground (concave notches). Min 3 cells long.
- **bridge**: 2x2 blocks in void connecting floating islands. Cannot overlap softEdge.
- **rail**: Closed loop on ground/bridge. Requires solid area >= 6x6. Cannot overlap other layers.
- **static**: 2x2 blocks on ground. Min 5x5 forbidden zone around doors. Blocks cannot touch each other. Must preserve door connectivity.
- **chaser**: Melee enemies. 0-3 cells from main path, prefer low squishy score. Cannot overlap static/bridge/rail/zoner.
- **zoner**: Area control enemies. **2x2 blocks** (1x1 only as a fallback where no 2x2 site fits), so one zoner = 4 cells. 0-5 cells from main path, prefer high squishy score. Every cell of a block must satisfy the constraints; cannot overlap static/bridge/rail/chaser. Distinct blocks cannot touch in any of the 8 directions.
- **dps**: Ranged damage enemies. 0-4 cells from main path, prefers proximity to chaser/static. Cannot overlap static/bridge/rail/zoner. **May share a cell with chaser** (ORT-122) — different categories, so it is not an ORT-93 spacing violation; only DPS-on-DPS adjacency is forbidden.
- **mobAir**: Air mobs. No ground requirement and **no overlap rule** (ORT-121) — neither validator checks it and hand-authored rooms stack it on other entities freely. **Centre-seeded and evenly distributed outward**, deterministically — the cell nearest the room centre is always taken, and the rest fill a centre-anchored grid. Zoner/chaser density only breaks ties within one grid slot. Door distance >= 4, edge distance >= 2, spacing >= 1. Generation keeps it clear of the other layers as a dispersion preference, not as a rule.

### Stage Type Rules

Counts are **spawns, not cells** (a zoner is a 2×2 block). `static` is counted in
2×2 blocks. When a stage type is supplied it overrides the request's counts.

| Stage | DPS | Chaser | Zoner | MobAir | Static | Notes |
|-------|-----|--------|-------|--------|--------|-------|
| start | 0 | 0 | 0 | 0 | 0 | Right door only (remaps to OZX Top) |
| teaching | 4-6 | 2-6 | 1-2 | 6 | 2-9 | |
| building | 6-9 | 6-10 | 1-2 | 6-9 | 2-9 | |
| pressure | 8-12 | 8-10 | 2-5 | 6-12 | 2-9 | Not bridge |
| peak | 8-12 | 8-10 | 2-3 | 12-18 | 2-9 | Full only; mobAir exceeds a 16×10 room's capacity, expect an ORT-110 warning |
| release | 2-4 | 2 | 1 | 6 | 2-9 | Minimal |
| boss | 0 | 0 | 0 | 0 | 0 | 6x6 center clear, max 2 doors |

Ranges last recalibrated in ORT-123/124/125 against the hand-authored rooms in
`Assets/StreamingAssets/TilemapData/normal`. No stage constrains room size
(ORT-109), but a small room may under-place — check the response's `warnings`.

## Error Handling

If the API returns an error:
- Show the HTTP status and error message
- Common issues: backend not running (connection refused), room too small for requested features, fewer than 2 doors, invalid roomCategory

Generation failures are `400 Generation failed` with the reason in the body.
A `main path: ...` reason means the room's doors could not be connected
(ORT-106) — either a doorway with no walkable cell on its wall, or two doors
with no walkable route between them:

```
main path: door left at (0,5): no walkable cell on that wall, the doorway is sealed
main path: doors left (0,5) and right (19,5) are not connected by walkable ground
```

These should never occur in normal operation — `ensureDoorsWalkable` guarantees
the invariant. Treat one as a generator bug worth reporting, not a bad request;
retrying the same parameters is not a fix.
