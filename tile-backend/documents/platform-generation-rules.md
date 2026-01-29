# Platform Room Auto-Generation Rules

This document describes the auto-generation algorithm for platform-type rooms.

## Input Parameters

| Parameter | Description |
|-----------|-------------|
| `width` | Room width (10-200) |
| `height` | Room height (10-200) |
| `doors` | Doors to connect (at least 2 required: top, right, bottom, left) |
| `softEdgeCount` | Suggested number of soft edges to place (optional, default 0) |
| `staticCount` | Suggested number of statics to place (optional, default 0) |
| `turretCount` | Suggested number of turrets to place (optional, default 0) |
| `mobGroundCount` | Suggested number of mob ground to place (optional, default 0) |
| `mobAirCount` | Suggested number of mob air (fly) to place (optional, default 0) |

## Ground Layer Generation

Platform rooms generate ground in two main phases:
1. **Platform Generation** - Create large platforms using one of two strategies
2. **Eraser Operations** - Create void areas within the platforms

### Strategy Selection

The algorithm automatically selects between two strategies:

| Strategy | Condition | Description |
|----------|-----------|-------------|
| Strategy 1: Center Platform | Default or when doors don't form corner groups | Single large center platform with door connections |
| Strategy 2: Corner Groups | When doors can form corner pairs (e.g., top+left) | Multiple platforms, one per corner group |

**Corner group detection**: Strategy 2 is available when doors can be grouped into corner pairs:
- Top-Left: top + left doors
- Top-Right: top + right doors
- Bottom-Left: bottom + left doors
- Bottom-Right: bottom + right doors

When Strategy 2 is available, there's a 50% chance it will be selected.

### Strategy 1: Center Platform

1. **Generate large platform in center**:
   - Width: Random between `width/2 + 1` and `width - 4`
   - Height: Random between `height/2 + 1` and `height - 4`
   - Position: Centered in the room

2. **Connect all doors**:
   - Brush size: Randomly select from 2×2, 3×3, or 4×4
   - Path type (50/50 chance):
     - **Via center**: Door → room center → platform
     - **Direct**: Door → nearest platform edge

```
Example Strategy 1 output (20×15, doors: top, bottom, left, right):

····████████████····
····████████████····
····████████████····
████████████████████
████████████████████
████████████████████
████████████████████
████████████████████
████████████████████
████████████████████
████████████████████
····████████████····
····████████████····
····████████████····
····████████████····

█ = Ground (platform)
· = Void
```

### Strategy 2: Corner Groups

1. **Group doors into corner pairs**:
   - Find all valid corner groups from connected doors
   - Example: doors [top, left, bottom, right] → groups: top-left, top-right, bottom-left, bottom-right

2. **For each corner group**:
   - Generate a large platform (width > room_width/2, height > room_height/2)
   - Position platform anchored to the corresponding corner
   - Connect the two doors in the group to the platform

```
Example Strategy 2 output (20×20, doors: top, left):

████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
████████████████····
····················
····················
····················
····················
····················
····················
····················
····················

Platform anchored to top-left corner, connecting top and left doors.
```

## Eraser Operations

After platform generation, eraser operations create void areas within the ground.

### Eraser Constraints

**IMPORTANT**: Each eraser operation must preserve door connectivity. If an operation would disconnect any doors, it is **rolled back**.

### Eraser Methods

| Method | Description | Brush Sizes |
|--------|-------------|-------------|
| `center_single` | Single void area in room center | 2×2, 3×3, 3×4, 4×4, 4×5 |
| `center_symmetric_2` | Two symmetric void areas (left-right or top-bottom) | 2×2, 3×3, 3×4 |
| `center_symmetric_3` | Three void areas (center + two symmetric) | 2×2, 3×3, 3×4 |
| `corners` | Erase platform corners (Strategy 1 only) | 2×2, 3×3, 3×4 |
| `unconnected_door_direction` | Void towards unconnected door direction (Strategy 1 only) | 3×3, 3×4, 4×4 |
| `strategy2_corner` | Erase random platform corner (Strategy 2 only) | 2×2, 3×3 |

### Corner Eraser Probability (Strategy 1)

When using the `corners` method:
- 1 corner: 5% probability
- 2 corners: 20% probability
- 3 corners: 5% probability
- 4 corners: 70% probability

### Eraser Execution Steps

1. **Select erase count**: Random between 0 and 3
2. **For each erase**:
   - Select a random method (no repeats)
   - Apply the eraser
   - Check door connectivity
   - If connectivity broken → rollback (but count still consumed)
3. **Continue** until count exhausted or no methods remaining

```
Example with eraser operations:

Before erasing:            After center_single (3×3):
████████████████████      ████████████████████
████████████████████      ████████████████████
████████████████████      ████████████████████
████████████████████      ████████████████████
████████████████████      ████████████████████
████████████████████      ████████████████████
████████████████████      ████████···█████████
████████████████████      ████████···█████████
████████████████████      ████████···█████████
████████████████████      ████████████████████

After corners (4 corners, 2×2):
··██████████████████
··██████████████████
████████████████··██
████████████████··██
████████████████████
████████████████████
████████···█████████
████████···█████████
██████████████████··
██████████████████··
```

## Other Layers

After ground generation, the following layers are generated using the same algorithms as bridge rooms:

| Layer | Description |
|-------|-------------|
| SoftEdge | Fills concave void notches in ground |
| Bridge | Connects floating islands and fills concave gaps |
| Static | 2×2 obstacle blocks on ground |
| Turret | 1×1 turrets preferring ground corners |
| MobGround | Ground-based mob spawn points |
| MobAir | Flying mob spawn points |

## API Endpoint

```
POST /api/v1/generate/platform
```

### Request Body

```json
{
  "width": 20,
  "height": 20,
  "doors": ["top", "bottom", "left", "right"],
  "softEdgeCount": 3,
  "staticCount": 4,
  "turretCount": 3,
  "mobGroundCount": 5,
  "mobAirCount": 4
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
    "turret": [[...]],
    "mobGround": [[...]],
    "mobAir": [[...]],
    "doors": {
      "top": 1,
      "right": 1,
      "bottom": 1,
      "left": 1
    },
    "roomType": "platform",
    "meta": {
      "name": "platform-20x20",
      "version": 1,
      "width": 20,
      "height": 20
    }
  },
  "debugInfo": {
    "ground": {
      "strategy": "strategy1_center_platform",
      "platforms": [
        {
          "position": "(2,2)",
          "size": "16x16",
          "group": ""
        }
      ],
      "doorConnections": [
        {
          "from": "top (10,0)",
          "to": "platform",
          "pathType": "via center",
          "brushSize": "3x3"
        }
      ],
      "eraserOps": [
        {
          "method": "center_single",
          "position": "(8,8)",
          "size": "4x4",
          "rolledBack": false
        },
        {
          "method": "corners",
          "position": "[(0,0), (18,0), (0,18), (18,18)]",
          "size": "2x2",
          "rolledBack": false
        }
      ]
    },
    "softEdge": { ... },
    "bridgeLayer": { ... },
    "static": { ... },
    "turret": { ... },
    "mobGround": { ... },
    "mobAir": { ... }
  }
}
```

## Comparison with Bridge Rooms

| Aspect | Bridge Room | Platform Room |
|--------|-------------|---------------|
| Ground coverage | Low (paths between doors) | High (large platforms) |
| Primary structure | Paths connecting doors | Large rectangular platforms |
| Void areas | Surrounding the paths | Created by eraser within platforms |
| Strategy selection | Single strategy | Two strategies (center vs corner groups) |
| Typical use | Narrow corridor rooms | Open arena rooms |

## Visual Example

```
Platform room with all features (25×20):

     0         1         2
     012345678901234567890123456

 0:  ··███████████████████···
 1:  ··███████████████████···
 2:  ··█████████···█████████·
 3:  ████████████···█████████
 4:  █▓▓██████████████████▓▓█
 5:  █▓▓██████████████████▓▓█
 6:  ████████████████████████
 7:  █████T██████████████T███
 8:  ████████████████████████
 9:  ████████MM██████████████
10:  ████████MM██████████████
11:  ████████████████████████
12:  █████████████A██████████
13:  ████████████████████████
14:  ██▓▓████████████████▓▓██
15:  ██▓▓████████████████▓▓██
16:  ████████████████████████
17:  ··███████████████████···
18:  ··███████████████████···
19:  ··███████████████████···

█ = Ground (walkable)
▓ = Static (2×2 blocks)
T = Turret (1×1)
M = Mob Ground (spawn point)
A = Mob Air (flying spawn)
· = Void
```
