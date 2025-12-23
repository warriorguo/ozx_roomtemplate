# 5-Layer Tile Template Editor

A React-based visual editor for creating game room tile templates with 5 distinct layers and rule-based validation.

## Features

- **5-Layer System**: Ground, Static, Turret, MobGround, MobAir with hierarchical constraints
- **Paint/Erase Tools**: Intuitive editing with click and drag support
- **Real-time Validation**: Immediate rule checking with visual error highlighting
- **Ground Auto-Generation**: Rule-driven room layout generation
- **Export Validation**: Enforced validation at export time
- **Visual Error Feedback**: Red borders for constraint violations

## Quick Start

```bash
npm install
npm run dev
```

Open your browser and navigate to the displayed URL to start editing templates.

## Layer System

### Layer Hierarchy (Bottom to Top)

1. **Ground (地面)** - Foundation layer
   - **Values**: `0` (non-walkable) | `1` (walkable)
   - **Rules**: No constraints
   - **Purpose**: Defines basic room layout and walkable areas

2. **Static (静态物品)** - Static object placement zones
   - **Values**: `0` (no items) | `1` (item placement allowed)
   - **Rules**: `static=1` requires `ground=1`
   - **Purpose**: Areas where static objects can be placed

3. **Turret (炮塔)** - Defensive turret positions
   - **Values**: `0` (no turret) | `1` (turret position)
   - **Rules**: `turret=1` requires `ground=1 AND static=0`
   - **Purpose**: Strategic defensive positions

4. **MobGround (地面怪)** - Ground-based mob spawn points
   - **Values**: `0` (no spawn) | `1` (spawn point)
   - **Rules**: `mobGround=1` requires `ground=1 AND static=0 AND turret=0`
   - **Purpose**: Enemy spawn locations on ground

5. **MobAir (飞行怪)** - Air-based mob spawn points
   - **Values**: `0` (no spawn) | `1` (spawn point)
   - **Rules**: No constraints (can spawn anywhere)
   - **Purpose**: Flying enemy spawn locations

## Rule Validation

### Constraint Logic

The system enforces hierarchical constraints:

```typescript
// Validation rules for each layer
Static:    valid = static==0    OR ground==1
Turret:    valid = turret==0    OR (ground==1 AND static==0)
MobGround: valid = mobGround==0 OR (ground==1 AND static==0 AND turret==0)
MobAir:    valid = true // No constraints
```

### Error Highlighting

- **Red Border**: Cells with `value=1` that violate constraints
- **Real-time**: Validation updates immediately after edits
- **Export Block**: Invalid templates cannot be exported
- **Error Summary**: Detailed error list with locations and reasons

## Editing System

### Toggle-Based Editing

- **Click to Toggle**: Direct click on cells switches between `0` and `1`
  - Cell with `0` → Click → becomes `1`
  - Cell with `1` → Click → becomes `0`
- **Drag Operation**: Hold and drag to batch edit multiple cells
  - Starting cell determines drag mode
  - Drag from `0` cell → paints `1` on all dragged cells
  - Drag from `1` cell → erases to `0` on all dragged cells

### Layer Selection

- Click layer headers to switch active editing layer
- Only active layer responds to editing interactions
- Visual feedback shows which layer is active

## Ground Auto-Generation

### Room Types

1. **Rectangular**: Standard rectangular rooms with configurable wall thickness
2. **Cross**: Cross-shaped rooms for intersections
3. **Custom**: Start with empty ground for manual design

### Door System

- **Position**: Specify X/Y coordinates
- **Direction**: North, South, East, West
- **Auto-placement**: Doors automatically create walkable paths through walls

### Generation Process

```typescript
const roomSpec: RoomSpec = {
  width: 15,
  height: 11,
  roomType: "rectangular",
  wallThickness: 1,
  doorPositions: [
    { x: 7, y: 0, direction: "north" },
    { x: 14, y: 5, direction: "east" }
  ]
};
```

## Data Format

Templates are exported as JSON with 5 layers:

```json
{
  "version": 1,
  "width": 15,
  "height": 11,
  "ground": [[0,1,0,...], ...],
  "static": [[0,0,1,...], ...],
  "turret": [[0,0,0,...], ...],
  "mobGround": [[0,0,0,...], ...],
  "mobAir": [[0,1,0,...], ...]
}
```

## UI Layout

### Vertical Layer Stack
- Each layer has its own editing grid
- Ground layer includes auto-generation panel
- Layers stacked vertically for easy comparison
- Right sidebar shows template info and validation status

### Status Indicators
- **Valid/Invalid**: Global validation status
- **Error Count**: Number of constraint violations
- **Active Layer**: Current editing layer
- **Hovered Cell**: Real-time cell information

## Controls & Shortcuts

- **Layer Selection**: Click layer headers to switch editing layer
- **Cell Editing**: Click cells to toggle `0 ↔ 1`, drag for batch editing
- **Visibility Toggle**: Eye icons to show/hide layers
- **Error Display**: Toggle error highlighting
- **Export**: Only enabled when template is valid

## Example Workflow

1. **Start**: Create new template with desired dimensions
2. **Ground**: Use auto-generator or manually paint walkable areas
3. **Static**: Paint item placement zones on walkable ground
4. **Turret**: Place defensive positions (avoiding static areas)
5. **MobGround**: Set enemy spawns (avoiding static/turret areas)
6. **MobAir**: Place flying enemy spawns (no restrictions)
7. **Validate**: Fix any red-highlighted errors
8. **Export**: Save validated template as JSON

## Sample Template

A sample 15×11 template is included in `sample-5layer-template.json` demonstrating:
- Rectangular room with walls
- Strategic static item placement
- Turret positioning for defense
- Balanced mob spawn distribution
- Proper constraint adherence across all layers

## Development Notes

Built with:
- React + TypeScript
- Zustand for state management
- Rule-based validation system
- Real-time constraint checking
- Hierarchical layer architecture

This system is designed for design-time template creation, not runtime map generation.