# Full Room Auto-Generation Rules

This document describes the auto-generation algorithm for full-type rooms.

## Input Parameters

| Parameter | Description |
|-----------|-------------|
| `width` | Room width (4-200) |
| `height` | Room height (4-200) |
| `doors` | Doors to connect (at least 2 required: top, right, bottom, left) |
| `softEdgeCount` | Suggested number of soft edges to place (optional, default 0) |
| `staticCount` | Suggested number of statics to place (optional, default 0) |
| `chaserCount` | Suggested number of chasers to place (optional, default 0) |
| `zonerCount` | Suggested number of zoners to place (optional, default 0) |
| `dpsCount` | Suggested number of DPS enemies to place (optional, default 0) |
| `mobAirCount` | Suggested number of mob air (fly) to place (optional, default 0) |
| `stageType` | Stage type identifier (optional) |

## Ground Layer Generation

Full rooms start with **all tiles filled** as ground, then optionally carve out corners and center pits.

### Step 1: Fill All Ground

Set every cell in the ground layer to `1` (walkable).

```
Example 10×8 room after step 1:

██████████
██████████
██████████
██████████
██████████
██████████
██████████
██████████

█ = Ground (walkable)
```

### Step 2: Corner Erase (40% probability)

With **40% probability**, erase corners from the filled ground.

If the 40% check fails, skip directly to Step 3.

#### 2.1 Choose Brush

Two brush types, each with **50% probability**:

| Brush Type | Width Range | Height Range |
|------------|-------------|--------------|
| Horizontal (Brush 1) | `[1, M/2]` | `[1, 2]` |
| Vertical (Brush 2) | `[1, 2]` | `[1, N/2]` |

Where M = room width, N = room height. The actual brush size is randomly selected within these ranges.

#### 2.2 Select Corner Combination

Select one combination by weighted probability:

| Combination | Probability |
|-------------|-------------|
| [TL, BL, TR, BR] (all four corners) | 50% |
| [TL, BL] (left side) | 10% |
| [TR, BR] (right side) | 10% |
| [BL, TR] (diagonal) | 10% |
| [TL, BR] (diagonal) | 10% |
| [TL] (single corner) | 2.5% |
| [BL] (single corner) | 2.5% |
| [TR] (single corner) | 2.5% |
| [BR] (single corner) | 2.5% |

#### 2.3 Apply Corner Erasing

For each corner in the selected combination:

1. Position the brush at the corner (top-left of brush aligned to corner)
2. Erase the rectangle (set ground=0)
3. **Check door connectivity** using BFS
4. If connectivity is broken:
   - **Rollback** the erase
   - **Retry once** with a halved brush size (width/2, height/2, minimum 1)
   - If retry also breaks connectivity, **rollback and skip all remaining corners**

```
Example after erasing all four corners with a 3×2 brush:

···████···
██████████
██████████
██████████
██████████
██████████
██████████
···████···

· = Erased corner
█ = Ground
```

### Step 3: Center Pits (30% probability)

With **30% probability**, dig symmetric pits in the ground.

If the 30% check fails, skip this step.

#### 3.1 Choose Brush

| Parameter | Range |
|-----------|-------|
| Width | `[1, M/3]` |
| Height | `[1, N/2]` |

#### 3.2 Select Pit Count

Randomly select **1 to 4** pits.

#### 3.3 Select Symmetry

50/50 choice between:
- **Left-right symmetric**: Pits are mirrored across the vertical center line
- **Top-bottom symmetric**: Pits are mirrored across the horizontal center line

#### 3.4 Place Pits

For each pit:
1. Generate a random offset from center
2. Place the pit and its mirror counterpart
3. **Check door connectivity**
4. If connectivity is broken, **rollback the entire pit pair**

```
Example with 2 left-right symmetric pits:

███████████████████
███████████████████
███████████████████
██████···███···████
██████···███···████
███████████████████
███████████████████
███████████████████

· = Center pits (symmetric around center)
█ = Ground
```

## Other Layers

After ground generation, the following layers are generated using the same algorithms as bridge and platform rooms:

| Layer | Description |
|-------|-------------|
| SoftEdge | Fills concave void notches in ground |
| Bridge | Connects floating islands and fills concave gaps |
| Static | 2×2 obstacle blocks on ground |
| Chaser | 1×1 chaser enemies preferring ground corners |
| Zoner | 2×2/1×1 area-denial enemies |
| DPS | 1×1 damage-dealing enemies |
| MobAir | Flying mob spawn points |
| MainPath | Main traversal path through the room |

See [bridge-generation-rules.md](bridge-generation-rules.md) for detailed rules on each layer, and [enemy-system-rules.md](enemy-system-rules.md) for the enemy system.

## API Endpoint

```
POST /api/v1/generate/fullroom
```

### Request Body

```json
{
  "width": 20,
  "height": 20,
  "doors": ["top", "bottom", "left", "right"],
  "softEdgeCount": 3,
  "staticCount": 4,
  "chaserCount": 3,
  "zonerCount": 5,
  "dpsCount": 2,
  "mobAirCount": 4,
  "stageType": ""
}
```

### Response

```json
{
  "payload": {
    "ground": [[...]],
    "softEdge": [[...]],
    "bridge": [[...]],
    "static": [[...]],
    "chaser": [[...]],
    "zoner": [[...]],
    "dps": [[...]],
    "mobAir": [[...]],
    "mainPath": [[...]],
    "doors": {
      "top": 1,
      "right": 1,
      "bottom": 1,
      "left": 1
    },
    "stageType": "",
    "roomType": "full",
    "meta": {
      "name": "full-20x20",
      "version": 1,
      "width": 20,
      "height": 20
    }
  },
  "debugInfo": {
    "ground": {
      "cornerErase": {
        "skipped": false,
        "brushType": "horizontal",
        "brushSize": "3x1",
        "combo": "[TL,BL,TR,BR]",
        "corners": [
          {
            "corner": "top-left",
            "position": "(0,0)",
            "size": "3x1",
            "rolledBack": false
          }
        ]
      },
      "centerPits": {
        "skipped": false,
        "brushSize": "2x3",
        "pitCount": 2,
        "symmetry": "left-right",
        "pits": [
          {
            "position": "(5,8)",
            "size": "2x3",
            "rolledBack": false
          },
          {
            "position": "(13,8)",
            "size": "2x3",
            "rolledBack": false
          }
        ]
      }
    },
    "softEdge": { ... },
    "bridgeLayer": { ... },
    "static": { ... },
    "chaser": { ... },
    "zoner": { ... },
    "dps": { ... },
    "mobAir": { ... }
  }
}
```

## Comparison with Other Room Types

| Aspect | Full Room | Bridge Room | Platform Room |
|--------|-----------|-------------|---------------|
| Ground coverage | Very high (starts 100%) | Low (paths between doors) | High (large platforms) |
| Primary structure | Full fill with carved corners/pits | Paths connecting doors | Large rectangular platforms |
| Void areas | Corners and center pits | Surrounding the paths | Created by eraser within platforms |
| Carving operations | Corner erase + center pits | N/A | Eraser operations |
| Typical use | Large open rooms with minor obstacles | Narrow corridor rooms | Open arena rooms |
| Min dimensions | 4×4 | 4×4 | 10×10 |
