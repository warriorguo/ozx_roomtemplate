# Bridge Room Auto-Generation Rules

This document describes the auto-generation algorithm for bridge-type rooms.

## Input Parameters

| Parameter | Description |
|-----------|-------------|
| `size` | Room dimensions (M × N) |
| `doors` | Doors to connect (at least 2 required) |

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

## Other Layers

| Layer | Default Value |
|-------|---------------|
| SoftEdge | All 0 |
| Bridge | All 0 |
| Static | All 0 |
| Turret | All 0 |
| MobGround | All 0 |
| MobAir | All 0 |

## Visual Example

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
