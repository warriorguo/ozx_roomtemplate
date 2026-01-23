# Bridge Room Auto-Generation Rules

This document describes the auto-generation algorithm for bridge-type rooms.

## Input Parameters

| Parameter | Description |
|-----------|-------------|
| `width` | Room width (4-200) |
| `height` | Room height (4-200) |
| `doors` | Doors to connect (at least 2 required: top, right, bottom, left) |
| `softEdgeCount` | Suggested number of soft edges to place (optional, default 0) |
| `staticCount` | Suggested number of statics to place (optional, default 0) |
| `turretCount` | Suggested number of turrets to place (optional, default 0) |
| `mobGroundCount` | Suggested number of mob ground to place (optional, default 0) |
| `mobAirCount` | Suggested number of mob air (fly) to place (optional, default 0) |

## Ground Layer Generation

### Step 1: Door Connectivity

1. Randomly select a brush size from: **2×2**, **3×3**, **4×4**
2. Connect two doors using either:
   - **Direct line**: Straight path between doors
   - **L-shaped path through center**: Path goes from door → room center point → other door
3. Repeat until all required doors are connected

### Step 2: Draw Small Platforms

#### 2.1 Initialize Strategies

Initialize the following probability-weighted strategies and randomly select a draw count (1-3):

| Strategy | Weight | Condition |
|----------|--------|-----------|
| Screen center point | 50 | Always available |
| Left & right door positions | 10 | If left-right connected |
| Midpoints between center and left/right doors | 10 | If left-right connected |
| Top & bottom door positions | 10 | If top-bottom connected |
| Midpoints between center and top/bottom doors | 10 | If top-bottom connected |
| All connected door positions | 10 | Always available |
| Midpoints from center to all connected doors | 10 | Always available |

#### 2.2 Select Strategy by Weight

Select a strategy based on probability weights, which returns a set of points.

#### 2.3 Draw Platforms

Randomly select a brush from: **2×2**, **2×3**, **3×3**, **3×2**, **4×3**, **3×4**, **4×4**, **4×5**, **5×4**, **5×5**

For each point in the selected strategy's point set, draw a rectangle centered on that point using the brush.

**Mirror Symmetry Rules:**
- When drawing on Y-axis (vertical center line, e.g., top-bottom strategies): Mirror **left-right** to maintain symmetry
- When drawing on X-axis (horizontal center line, e.g., left-right strategies): Mirror **top-bottom** to maintain symmetry
- Center and "all doors/midpoints" strategies: No mirroring (already symmetric or covers all positions)

#### 2.4 Repeat or Exit

- Decrement draw count
- Remove the used strategy from the available strategies
- If draw count = 0 or no strategies remain, exit
- Otherwise, repeat from step 2.2

### Step 3: Floating Islands (Optional)

After drawing platforms, optionally add floating islands in empty void areas:

#### 3.1 Find Empty Areas

Scan the ground layer to find empty rectangular areas (void regions) that are at least **4×4** cells.

#### 3.2 Island Placement Loop

For each empty area (in random order):

1. **50% probability check**: With 50% probability, skip this area
2. **Find valid position**: Search for a position within the empty area where an island can be placed:
   - Island size: Random between **2×2** and the maximum that fits
   - **Distance constraint**: Island must be exactly **2 cells away** from existing ground (not closer, not farther)
   - The 2-cell gap ensures islands are visually separated but still close enough to be bridged
3. **Place island**: If valid position found, draw the island (set ground=1)
4. **Continue or stop**: Continue to next area, or stop if 50% probability check fails

#### 3.3 Distance Constraint Details

An island position is valid if:
- **Margin check**: No existing ground within 2 cells of the island (inner margin)
- **Outer ring check**: Existing ground must exist at exactly 3 cells distance (to ensure island is close enough to be connected by bridges)

```
Example valid island placement:

     0123456789
y 0: ··████████   ← Main ground
y 1: ··████████
y 2: ··████████
y 3: ··········   ← 2-cell gap (void)
y 4: ··········   ← 2-cell gap (void)
y 5: ····██····   ← Floating island (2×2)
y 6: ····██····
y 7: ··········

The island at (4,5)-(5,6) is exactly 2 cells below the ground at y=2.
```

## Soft Edge Layer Generation

### Input Parameters

| Parameter | Description |
|-----------|-------------|
| `softEdgeCount` | Suggested number of soft edges to place (optional, default 0) |

### Placement Rules

1. **Soft Edge shape**: 1×N or N×1 strips where N > 2 (minimum 3 cells)
2. **Void placement**: Soft edges are placed in void cells (ground=0), NOT on ground
3. **Concave placement**: Soft edges fill void "notches" that are surrounded by ground on multiple sides
4. **No overlap**: Cannot place in the same location twice
5. **Door distance**: Must be at least **2 cells** (Manhattan distance) away from any door

### Concave Area Definition (Void Notches)

A **concave area** (or **void notch**) is a strip of void cells that is surrounded by ground, forming a U-shaped depression:

- **Horizontal concave (1×N)**: Void cells where:
  - Ground exists above OR below for the entire strip (forming the "floor" or "ceiling" of the notch)
  - Ground exists on both left and right ends (closing the notch)
  - Void exists on the opposite horizontal side (the opening)

- **Vertical concave (N×1)**: Void cells where:
  - Ground exists to the left OR to the right for the entire strip (forming the "wall" of the notch)
  - Ground exists on both top and bottom ends (closing the notch)
  - Void exists on the opposite vertical side (the opening)

### Visual Example of Concave Areas

```
Ground layer showing concave notches:

     01234567890123456789
y 0: ·······██████·······
y 1: ·······██████·······
y 2: ·······██████·······
y 3: ███····██████···████   ← Concave notches at (3-6,3) and (13-15,3)
y 4: ████████████████████
y 5: ████████████████████
y 6: ████████████████████
y 7: ████████████████████
y 8: ███····██████···████   ← Concave notches at (3-6,8) and (13-15,8)
y 9: ····················

The notches at row 3: void cells (3,3)-(6,3) and (13,3)-(15,3)
  - Ground below (row 4)
  - Ground on left and right ends
  - Void above (rows 0-2)
  → These are valid horizontal concaves opening upward

The notches at row 8: void cells (3,8)-(6,8) and (13,8)-(15,8)
  - Ground above (row 7)
  - Ground on left and right ends
  - Void below (rows 9+)
  → These are valid horizontal concaves opening downward
```

### Placement Steps

1. **Find valid placements**: Scan all void positions to find horizontal and vertical concave notches that:
   - Are at least 3 cells long
   - Are closed on both ends by ground
   - Have ground on one horizontal/vertical side
   - Are far enough from doors
   - Don't overlap with existing soft edges
2. **Shuffle placements**: Randomize the order for variety
3. **Place until done**: Place soft edges until target count reached or all valid placements exhausted

### Soft Edge Layer Example

```
Ground layer with soft edges filling concave notches:

     01234567890123456789
y 0: ·······██████·······
y 1: ·······██████·······
y 2: ·······██████·······
y 3: ███SSSS██████SSS████   S = Soft Edge filling void notches
y 4: ████████████████████
y 5: ████████████████████
y 6: ████████████████████
y 7: ████████████████████
y 8: ███SSSS██████SSS████   S = Soft Edge filling void notches
y 9: ····················

█ = Ground (walkable)
S = Soft Edge (in void notch)
· = Empty void
```

## Static Layer Generation

### Input Parameters

| Parameter | Description |
|-----------|-------------|
| `staticCount` | Suggested number of statics to place (optional, default 0) |

### Placement Rules

1. **Static size**: Fixed at 2×2 cells per static
2. **Ground requirement**: All 4 cells of a static must be on ground (ground=1)
3. **Door avoidance**: Statics cannot be placed at or adjacent to door positions (5×5 forbidden zone around each door)
4. **No overlap with SoftEdge/Bridge**: Static cells must not overlap with softEdge or bridge layers
5. **No touching**: Statics cannot touch each other (including diagonals) - minimum 1 cell gap required
6. **Connectivity preservation**: After placing a static, all doors must remain connected via walkable paths

### Placement Strategies

Two alternating strategies are used:

| Strategy | Description |
|----------|-------------|
| Center Outward | Start from room center, prioritize positions closer to center |
| Edge Inward | Start from room edges, prioritize positions closer to edges |

### Placement Steps

1. Initialize valid positions list (all 2×2 positions satisfying ground, softEdge, bridge, and door constraints)
2. Set remaining count = staticCount
3. Select current strategy (alternates between Center Outward and Edge Inward)
4. Sort valid positions by current strategy
5. Attempt to place one static:
   - Find first valid position that maintains door connectivity
   - If found: place static, remove position, filter out touching positions, decrement remaining
   - If not found: increment failure counter
6. Switch strategy
7. Repeat from step 4 until remaining = 0 or max attempts reached

### Connectivity Check

Uses BFS (Breadth-First Search) to verify all doors remain connected after a hypothetical static placement. A static placement is rejected if it would block the only path between any two doors.

## Turret Layer Generation

### Input Parameters

| Parameter | Description |
|-----------|-------------|
| `turretCount` | Suggested number of turrets to place (optional, default 0) |

### Placement Rules

1. **Turret size**: Fixed at 1×1 cell per turret
2. **Ground requirement**: Turret cell must be on ground (ground=1)
3. **Door distance**: Turrets must be at least **4 cells** (Manhattan distance) away from any door
4. **Turret spacing**: Turrets must be at least **2 cells** (Manhattan distance) apart from each other
5. **No overlap**: Turret cells must not overlap with softEdge, bridge, or static layers
6. **Connectivity preservation**: After placing a turret, all doors must remain connected via walkable paths

### Placement Preference

Turrets are preferentially placed (in order of priority):
1. **Ground corners (90°/270°)**: Tiles where ground forms an L-shape (highest priority)
   - **90° right angle**: Ground tile with exactly 2 orthogonal neighbors that are adjacent (L-shape corner)
   - **270° inner corner**: Ground tile with exactly 3 orthogonal neighbors (inverted L-shape)
2. **Near room corners**: Within 2 cells of room corners
3. **Near room edges**: Within 2 cells of room edges
4. **Center outward**: Among valid positions, closer to center is preferred

### Placement Steps

1. Find all valid positions (satisfying ground, layer overlap, door distance constraints)
2. Sort positions by preference score (corners > edges > center distance)
3. Set remaining count = turretCount
4. Attempt to place one turret:
   - Find first valid position that maintains door connectivity
   - If found: place turret, filter out positions too close, decrement remaining
   - If not found: decrement max attempts
5. Repeat from step 4 until remaining = 0 or max attempts reached

### Connectivity Check

Uses BFS (Breadth-First Search) to verify all doors remain connected after a hypothetical turret placement. A turret placement is rejected if it would block the only path between any two doors.

## Mob Ground Layer Generation

### Input Parameters

| Parameter | Description |
|-----------|-------------|
| `mobGroundCount` | Suggested number of mob ground to place (optional, default 0) |

### Placement Rules

1. **Mob Ground size**: 2×2 (preferred) or 1×1 (fallback)
2. **Ground requirement**: All cells must be on ground (ground=1)
3. **Door distance**: Mob ground must be at least **2 cells** (Manhattan distance) away from any door
4. **No touching**: Mob ground cannot touch each other (including diagonals)
5. **No overlap**: Mob ground cells must not overlap with softEdge, bridge, static, or turret layers
6. **Size preference**: Always try 2×2 first, fall back to 1×1 if 2×2 cannot fit

### Placement Strategies

Three strategies are available, each group uses a different strategy:

| Strategy | Description |
|----------|-------------|
| Large Open Area | If there's a large walkable area (≥10% of room) connected to doors, place from its center outward |
| Near Doors | Place around door areas (but respecting minimum distance) |
| Center Outward | Place from room center outward |

### Placement Steps

1. **Divide into groups**: Split target count into 2-3 groups with roughly equal sizes
   - If any group has less than 1, merge groups until all have ≥1
   - Example: count=5 → groups=[2, 2, 1]

2. **Assign strategies**: Each group gets a unique strategy (no duplicates)
   - Strategies are shuffled randomly
   - Large Open Area strategy is skipped if no suitable area exists

3. **Execute per group**: For each group:
   - Find all valid positions based on rules
   - Sort positions by strategy preference
   - Place mob ground (prefer 2×2, fallback to 1×1)
   - Continue until group count reached or placement fails

### Mob Ground Layer Example

```
mobGroundCount: 4

Ground + Static + Turret + MobGround overlay:

····████████····████
····████████····████
····██▓▓████····████
····██▓▓███T····████
····████████····████
████████████████████
██T████MM███████▓▓██
████████MM██████▓▓██
████████████████████
██████████████M█████
····████████········
····███T████········
····████████········
····████████····T···
····████████········

█ = Ground only (walkable)
▓ = Static on ground (blocked, 2×2 blocks)
T = Turret on ground (blocked, 1×1 tile)
M = Mob Ground (2×2 or 1×1 spawn point)
· = Empty (void)

Note: Mob ground maintains minimum 2-cell distance from doors
      and cannot touch other mob ground (including diagonals).
```

## Mob Air (Fly) Layer Generation

### Input Parameters

| Parameter | Description |
|-----------|-------------|
| `mobAirCount` | Suggested number of mob air (fly) to place (optional, default 0) |

### Placement Rules

1. **Mob Air size**: Fixed at 1×1 cell
2. **No ground requirement**: Flying mobs can spawn anywhere (ground=0 or ground=1)
3. **Door distance**: Mob air must be at least **4 cells** (Manhattan distance) away from any door
4. **Edge distance**: Mob air must be at least **2 cells** away from room edges (all four sides)
5. **No touching**: Mob air cannot touch each other (including diagonals)
6. **No overlap**: Mob air cells must not overlap with softEdge, bridge, static, turret, or mobGround layers

### Placement Strategies

One strategy is randomly selected:

| Strategy | Description |
|----------|-------------|
| Center Outward | Place from room center outward, closest to center first |
| Evenly Spaced | Distribute mob air evenly across the map using grid-based selection calculated from target count |

### Placement Steps

1. **Select strategy**: Randomly choose between Center Outward or Evenly Spaced
2. **Find valid positions**: Collect all cells that satisfy placement rules
3. **Sort/arrange positions**:
   - Center Outward: Sort by distance from center (closest first)
   - Evenly Spaced: Calculate grid dimensions based on target count (cols × rows ≥ targetCount), then select nearest valid position to each grid cell center
4. **Place mob air**: Place 1×1 mob air at valid positions until target count reached

### Evenly Spaced Algorithm

For even distribution based on target count:
1. **Calculate grid dimensions**: Determine cols and rows such that `cols × rows ≥ targetCount`, maintaining aspect ratio similar to room dimensions
   - Example: For `targetCount=4` in a 20×20 room → 2×2 grid
   - Example: For `targetCount=6` in a 30×20 room → 3×2 or similar grid
2. **Calculate cell size**: `cellWidth = width / cols`, `cellHeight = height / rows`
3. **Find ideal positions**: For each grid cell, calculate its center point
4. **Select nearest valid position**: For each cell center, find the nearest valid position from the available positions
5. **Fill remaining**: If more positions needed, fill from remaining valid positions

### Mob Air Layer Example

```
mobAirCount: 4

Ground + Static + Turret + MobGround + MobAir overlay:

····████████····████
····████████····████
····██▓▓██A█····████
····██▓▓███T····████
····████████····████
████████████████████
██T████MM██A████▓▓██
████████MM██████▓▓██
████████████████████
██████████████M█████
····██A█████········
····███T████········
····████████····A···
····████████····T···
····████████········

█ = Ground only (walkable)
▓ = Static on ground (blocked, 2×2 blocks)
T = Turret on ground (blocked, 1×1 tile)
M = Mob Ground (2×2 or 1×1 spawn point)
A = Mob Air (1×1 flying mob spawn point)
· = Empty (void)

Note: Mob air maintains minimum 4-cell distance from doors,
      minimum 2-cell distance from room edges,
      and cannot touch other mob air (including diagonals).
```

## Bridge Layer Generation

The bridge layer connects floating islands to the main ground and fills concave gaps in the ground.

### Purpose

Bridges serve two purposes:
1. **Connect floating islands**: Ensure all isolated ground regions are reachable
2. **Fill concave gaps**: Add walkable paths in horizontal void notches within the ground

### Placement Rules

1. **Bridge size**: Fixed at **2×2** cells
2. **Void requirement**: All 4 cells must be on void (ground=0)
3. **Touch requirement**: Bridge must fully touch (2+ adjacent cells) both source and target
4. **No overlap**: Cannot overlap with existing bridges or soft edges

### Step 1: Connect Floating Islands

If the ground layer contains disconnected regions (islands):

1. **Find all islands**: Use flood-fill to identify connected ground regions
2. **Identify main ground**: The largest connected region is the "main ground"
3. **Connect each island**: For each unconnected island:
   - Find the nearest position where a 2×2 bridge can connect the island to main ground (or another connected island)
   - Bridge must touch 2+ cells of the island AND 2+ cells of the target
   - Place the bridge

### Step 2: Fill Concave Gaps

Even when all ground is connected, horizontal concave gaps can benefit from bridges:

1. **Find horizontal gaps**: Scan each row for void segments where:
   - Ground exists on both left and right sides
   - Gap width is at least **4 cells**
2. **Check concave condition**: The gap must have ground above it (≥50% coverage of gap width)
3. **Place bridge**: Place a 2×2 bridge centered in the gap

```
Example concave gap with bridge:

     01234567890123456789
y 7: ████████████████████   ← Full ground above
y 8: ███··········███████   ← Gap from x=3 to x=12
y 9: ····················   ← Void below

Gap at y=8 (x=3 to x=13, width=10):
- Ground on left (x=0-2) and right (x=13-19)
- Ground above (y=7, full row)
- Bridge placed at center: (7,8)-(8,9)

After bridge placement:
y 8: ███····BB····███████   B = Bridge (2×2)
y 9: ········BB··········
```

### Bridge Layer Debug Info

```json
{
  "bridgeLayer": {
    "skipped": false,
    "islandsFound": 3,
    "bridgesPlaced": 4,
    "connections": [
      {
        "from": "island (5,2)-(8,4)",
        "to": "main ground",
        "position": "(5,5)",
        "size": "2x2"
      }
    ],
    "concaveGapBridges": [
      {
        "from": "concave gap at y=8",
        "to": "x=3 to x=12",
        "position": "(7,8)",
        "size": "2x2"
      }
    ]
  }
}
```

## Other Layers

| Layer | Description |
|-------|-------------|
| SoftEdge | Generated based on `softEdgeCount` parameter (fills void notches) |
| Bridge | Connects floating islands and fills concave gaps (auto-generated) |

## Visual Examples

### Ground Layer Example

```
Doors: Top, Bottom, Left, Right
Size: 20×15

····████████····████
····████████····████
····████████····████
····████████····████
····████████····████
████████████████████
████████████████████
████████████████████
████████████████████
████████████████████
····████████········
····████████········
····████████········
····████████········
····████████········

█ = Ground (walkable)
· = Empty (void)
```

### Static Layer Example

```
staticCount: 2

Ground + Static overlay:

····████████····████
····████████····████
····██▓▓████····████
····██▓▓████····████
····████████····████
████████████████████
████████████████▓▓██
████████████████▓▓██
████████████████████
████████████████████
····████████········
····████████········
····████████········
····████████········
····████████········

█ = Ground only (walkable)
▓ = Static on ground (blocked, 2×2 blocks)
· = Empty (void)

Note: Statics maintain minimum 1-cell gap from each other
      and doors, preserving path connectivity.
```

### Turret Layer Example

```
turretCount: 4

Ground + Static + Turret overlay:

····████████····████
····████████····████
····██▓▓████····████
····██▓▓███T····████
····████████····████
████████████████████
██T█████████████▓▓██
████████████████▓▓██
████████████████████
████████████████████
····████████········
····███T████········
····████████········
····████████····T···
····████████········

█ = Ground only (walkable)
▓ = Static on ground (blocked, 2×2 blocks)
T = Turret on ground (blocked, 1×1 tile)
· = Empty (void)

Note: Turrets maintain minimum 4-cell distance from doors
      and minimum 2-cell distance from each other.
```

## API Response Debug Info

The API response includes a `debugInfo` field that provides detailed information about the generation process.

### Debug Info Structure

Each layer's debug info includes:
- `skipped`: Boolean indicating if generation was skipped (when count=0)
- `skipReason`: String explaining why generation was skipped (only when skipped=true)
- `targetCount`: Requested placement count
- `placedCount`: Actual number placed
- `placements`: Array of successful placements
- `misses`: Array of failed placement attempts with reasons

```json
{
  "payload": { ... },
  "debugInfo": {
    "ground": {
      "doorConnections": [
        {
          "from": "top (10,0)",
          "to": "bottom (10,19)",
          "pathType": "direct",
          "brushSize": "4x4"
        }
      ],
      "platforms": [
        {
          "strategy": "center",
          "brushSize": "6x6",
          "points": ["(10,10)"],
          "mirror": "none"
        }
      ],
      "floatingIslands": [
        {
          "position": "(4,10)",
          "size": "3x2",
          "fromArea": "(0,8) 8x6",
          "skipped": false
        },
        {
          "position": "",
          "size": "",
          "fromArea": "",
          "skipped": true,
          "skipReason": "stopped by 50% probability check"
        }
      ]
    },
    "bridgeLayer": {
      "skipped": false,
      "islandsFound": 2,
      "bridgesPlaced": 2,
      "connections": [
        {
          "from": "island (4,10)-(6,11)",
          "to": "main ground",
          "position": "(4,8)",
          "size": "2x2"
        }
      ],
      "concaveGapBridges": [
        {
          "from": "concave gap at y=15",
          "to": "x=5 to x=14",
          "position": "(8,15)",
          "size": "2x2"
        }
      ]
    },
    "softEdge": {
      "skipped": false,
      "targetCount": 3,
      "placedCount": 3,
      "placements": [
        {
          "position": "(3,3)",
          "size": "4x1",
          "reason": "ground concave area"
        }
      ],
      "misses": [
        {
          "reason": "overlapping with already placed soft edge",
          "count": 1
        }
      ]
    },
    "static": {
      "skipped": false,
      "targetCount": 3,
      "placedCount": 3,
      "placements": [
        {
          "position": "(5,5)",
          "size": "2x2",
          "reason": "strategy: center_outward, valid position with connectivity preserved"
        }
      ],
      "misses": [
        {
          "reason": "position would block door connectivity",
          "count": 5
        }
      ]
    },
    "turret": {
      "skipped": false,
      "targetCount": 4,
      "placedCount": 4,
      "placements": [
        {
          "position": "(15,3)",
          "size": "1x1",
          "reason": "ground corner (90° right angle)"
        }
      ]
    },
    "mobGround": {
      "skipped": false,
      "targetCount": 3,
      "placedCount": 3,
      "groups": [
        {
          "groupIndex": 0,
          "strategy": "center_outward",
          "targetCount": 1,
          "placedCount": 1,
          "placements": [
            {
              "position": "(8,8)",
              "size": "2x2",
              "reason": "preferred 2x2 placement via center_outward strategy"
            }
          ],
          "misses": []
        }
      ],
      "misses": [
        {
          "reason": "large_open_area strategy not viable (no 4x4 open area found)"
        }
      ]
    },
    "mobAir": {
      "skipped": false,
      "targetCount": 4,
      "placedCount": 4,
      "strategy": "evenly_spaced",
      "placements": [
        {
          "position": "(3,3)",
          "size": "1x1",
          "reason": "placed via evenly_spaced strategy (on void, flying mob)"
        }
      ]
    }
  }
}
```

### Skipped Generation Example

When a layer's count is 0 or not specified:

```json
{
  "softEdge": {
    "skipped": true,
    "skipReason": "softEdgeCount is 0 or not specified",
    "targetCount": 0,
    "placedCount": 0,
    "placements": null
  }
}
```

### Common Miss Reasons

| Layer | Possible Miss Reasons |
|-------|----------------------|
| Ground (Floating Islands) | `stopped by 50% probability check`, `no valid position found in empty area` |
| Soft Edge | `no valid concave areas found in ground layer`, `overlapping with already placed soft edge`, `only N valid placements available, needed M more` |
| Bridge | `no floating islands found (all ground is connected)`, `cannot find valid bridge path for island at (x,y)-(x,y)`, `no bridges needed (no floating islands and no concave gaps)` |
| Static | `no valid 2x2 positions found`, `position invalidated by previous placement (touching existing static)`, `position would block door connectivity`, `reached max strategy attempts` |
| Turret | `no valid positions found`, `position invalidated (too close to existing turret or blocked)`, `position would block door connectivity` |
| Mob Ground | `large_open_area strategy not viable (no 4x4 open area found)`, `group N skipped: no more strategies available`, `no valid positions available`, `positions found but neither 2x2 nor 1x1 placement possible` |
| Mob Air | `no valid positions found`, `position invalidated by previous placement (already occupied)`, `exhausted all N valid positions, needed M more` |

### Strategy Names

| Layer | Strategies |
|-------|------------|
| Ground Platforms | `center`, `left_right_doors`, `left_right_midpoints`, `top_bottom_doors`, `top_bottom_midpoints`, `all_doors`, `all_midpoints` |
| Floating Islands | N/A (probabilistic placement in empty areas) |
| Soft Edge | N/A (finds concave areas automatically) |
| Bridge | N/A (connects islands via flood-fill + fills concave gaps) |
| Static | `center_outward`, `edge_inward` (alternating) |
| Turret | Priority-based: `ground corner (90° right angle)`, `ground corner (270° inner corner)`, `near room corner`, `near room edge`, `center outward placement` |
| Mob Ground | `large_open_area`, `near_doors`, `center_outward` |
| Mob Air | `center_outward`, `evenly_spaced` |
