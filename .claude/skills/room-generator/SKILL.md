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
| `stageType` | string | `""` | Stage type: `"start"`, `"teaching"`, `"building"`, `"pressure"`, `"peak"`, `"release"`, `"boss"`, or empty (defaults to `"default"` in output). Controls enemy count ranges. |
| `roomCategory` | string | `"normal"` | Room category: `"normal"`, `"basement"`, `"test"`, `"cave"`. Passed through to output. |
| `softEdgeCount` | int | `3` | Number of soft edge strips to place in void notches |
| `railEnabled` | bool | `true` | Whether to generate a rail loop on the ground |
| `staticCount` | int | `8` | Number of 2x2 static obstacle blocks |
| `chaserCount` | int | `4` | Number of chaser placements (melee enemies near main path) |
| `zonerCount` | int | `2` | Number of zoner placements (area control enemies) |
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

**Before saving**, remap door directions for OZX Unity coordinate convention (transpose: top↔left, bottom↔right):

```python
DOOR_REMAP = {'top': 'left', 'right': 'bottom', 'bottom': 'right', 'left': 'top'}
BITMASK_REMAP = {1: 8, 2: 4, 4: 2, 8: 1}  # Top=1→Left=8, Right=2→Bottom=4, etc.

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
  }
}
```

### Layer Rules

- **ground**: Foundation layer. Full rooms start 100% filled then carve corners/pits. Bridge rooms connect doors with paths. Platform rooms use large rectangular blocks.
- **softEdge**: Placed in void cells adjacent to ground (concave notches). Min 3 cells long.
- **bridge**: 2x2 blocks in void connecting floating islands. Cannot overlap softEdge.
- **rail**: Closed loop on ground/bridge. Requires solid area >= 6x6. Cannot overlap other layers.
- **static**: 2x2 blocks on ground. Min 5x5 forbidden zone around doors. Blocks cannot touch each other. Must preserve door connectivity.
- **chaser**: Melee enemies. 0-3 cells from main path, prefer low squishy score. Cannot overlap static/bridge/rail/zoner.
- **zoner**: Area control enemies. 0-5 cells from main path, prefer high squishy score. Cannot overlap static/bridge/rail/chaser.
- **dps**: Ranged damage enemies. 0-4 cells from main path, prefers proximity to chaser/static. Cannot overlap bridge/rail/zoner.
- **mobAir**: Air mobs. No ground requirement. Prefers zoner/chaser dense areas, spacing >= 1. Cannot overlap other entity layers.

### Stage Type Rules

| Stage | DPS | Chaser | Zoner | MobAir | Notes |
|-------|-----|--------|-------|--------|-------|
| start | 0 | 0 | 0 | 0 | Left door only (remaps to OZX Top) |
| teaching | 2-3 | 0 | 0 | 0 | DPS only |
| building | 2-3 | 2-3 | 0 | 0 | DPS + Chaser |
| pressure | 4-6 | 6-8 | 1 | 2-4 | Not bridge |
| peak | 6-12 | 6-8 | 2-3 | 2-4 | Full only |
| release | 0-2 | 0-2 | 0-1 | 0-2 | Minimal |
| boss | 0 | 0 | 0 | 0 | 6x6 center clear, max 2 doors |

## Error Handling

If the API returns an error:
- Show the HTTP status and error message
- Common issues: backend not running (connection refused), room too small for requested features, fewer than 2 doors, invalid roomCategory
