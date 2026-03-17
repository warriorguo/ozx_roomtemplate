---
name: room-generator
description: Generate game room templates (full/bridge/platform) via backend API. Use this skill when the user wants to generate a room, create a room template, test room generation, or batch-generate rooms. Triggers on phrases like "generate a room", "create room template", "test fullroom generation", "generate bridge room", "make a platform room", or any room generation task.
---

# Room Generator

Generate tile-based game room templates by calling the backend generation API. Produces a multi-layer room with ground, softEdge, bridge, rail, static, turret, mobGround, and mobAir layers.

## Parameters

All parameters have sensible defaults. Show them to the user and let them modify before generating.

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `endpoint` | string | `http://localhost:8090/api/v1` | Backend API base URL |
| `roomType` | string | `"full"` | Room type: `"full"` (ground almost fully filled), `"bridge"` (narrow corridor paths), or `"platform"` (large platform blocks) |
| `width` | int | `20` | Room width in tiles (4-200) |
| `height` | int | `12` | Room height in tiles (4-200) |
| `doors` | string[] | `["top","right","bottom","left"]` | Doors to connect. At least 2 required. Options: `top`, `right`, `bottom`, `left` |
| `softEdgeCount` | int | `3` | Number of soft edge strips to place in void notches |
| `railEnabled` | bool | `true` | Whether to generate a rail loop on the ground |
| `staticCount` | int | `8` | Number of 2x2 static obstacle blocks |
| `turretCount` | int | `4` | Number of 1x1 turret placements |
| `mobGroundCount` | int | `8` | Number of ground mob spawn points |
| `mobAirCount` | int | `10` | Number of air mob spawn points |
| `outputPath` | string | *(ask user)* | File path to save the full JSON response. If not provided, ask the user where to save it. |

## Workflow

### Step 1: Show parameters and confirm

Display the current parameter values in a table. Ask the user if they want to change anything before generating. Example:

```
Room Generation Parameters:
  endpoint:       http://localhost:8090/api/v1
  roomType:       full
  width:          20
  height:         12
  doors:          [top, right, bottom, left]
  softEdgeCount:  3
  railEnabled:    true
  staticCount:    8
  turretCount:    4
  mobGroundCount: 8
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
  "softEdgeCount": <softEdgeCount>,
  "railEnabled": <railEnabled>,
  "staticCount": <staticCount>,
  "turretCount": <turretCount>,
  "mobGroundCount": <mobGroundCount>,
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
| `T` | turret | 3 |
| `M` | mobGround | 4 |
| `A` | mobAir | 5 |
| `E` | softEdge | 6 |
| `B` | bridge | 7 |
| `█` | ground (walkable) | 8 |
| `·` | void | 9 |

Use this python snippet to render:
```python
import json, sys
data = json.load(sys.stdin)
p = data['payload']
h, w = len(p['ground']), len(p['ground'][0])
print(f"Room: {w}x{h} ({data['payload'].get('roomType','unknown')})")
print()
for y in range(h):
    line = ''
    for x in range(w):
        if p.get('rail') and p['rail'][y][x]: line += 'R'
        elif p['static'][y][x]: line += 'S'
        elif p['turret'][y][x]: line += 'T'
        elif p['mobGround'][y][x]: line += 'M'
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
- Turret: target vs placed count
- MobGround: target vs placed count
- MobAir: target vs placed count

### Step 5: Save the result

Ask the user for the output path if not specified. Save the **full JSON response** (payload + debugInfo) to the file. Use `python3 -c` or `Write` tool to save.

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
    "turret": [[0,1,...], ...],      // 1x1 turret positions
    "mobGround": [[0,1,...], ...],   // Ground mob spawn points
    "mobAir": [[0,1,...], ...],      // Air mob spawn points (no ground required)
    "doors": {
      "top": 0|1,
      "right": 0|1,
      "bottom": 0|1,
      "left": 0|1
    },
    "roomType": "full"|"bridge"|"platform",
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
    "turret": { "skipped": false, "targetCount": 4, "placedCount": 4, ... },
    "mobGround": { "skipped": false, "targetCount": 8, "placedCount": 7, ... },
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
- **turret**: 1x1 on ground. Min 4 cells from doors. Min 2 cells between turrets.
- **mobGround**: 2x2 (preferred) or 1x1 on ground. Min 4 cells from doors.
- **mobAir**: 1x1 anywhere. Min 4 cells from doors. Min 2 cells from room edges.

## Error Handling

If the API returns an error:
- Show the HTTP status and error message
- Common issues: backend not running (connection refused), room too small for requested features, fewer than 2 doors
