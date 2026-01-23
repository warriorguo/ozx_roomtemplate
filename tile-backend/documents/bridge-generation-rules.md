# Bridge Room Auto-Generation Rules

This document describes the auto-generation algorithm for bridge-type rooms.

## Input Parameters

| Parameter | Description |
|-----------|-------------|
| `width` | Room width (4-200) |
| `height` | Room height (4-200) |
| `doors` | Doors to connect (at least 2 required: top, right, bottom, left) |
| `staticCount` | Suggested number of statics to place (optional, default 0) |
| `turretCount` | Suggested number of turrets to place (optional, default 0) |

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

## Other Layers

| Layer | Default Value |
|-------|---------------|
| SoftEdge | All 0 |
| Bridge | All 0 |
| MobGround | All 0 |
| MobAir | All 0 |

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
