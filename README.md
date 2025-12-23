# Room Template Editor

A React-based visual editor for creating game room tile templates with three distinct layers.

## Features

- **Visual Grid Editor**: Click-based editing with real-time visual feedback
- **Three-Layer System**: Ground, Static, and Monster layers with different rules
- **Import/Export**: JSON-based template files with validation and auto-correction
- **Layer Management**: Toggle visibility and switch between editing layers
- **Smart Validation**: Automatic constraint enforcement and error correction

## Quick Start

```bash
npm install
npm run dev
```

Open your browser and navigate to the displayed URL to start editing templates.

## Data Format

Templates are exported as JSON with the following structure:

```json
{
  "version": 1,
  "width": 15,
  "height": 11,
  "ground": [[0,1,0,...], ...],
  "static": [[0,0,1,...], ...],
  "monster": [[0,1,2,...], ...]
}
```

### Layer Rules

#### Ground Layer
- **Values**: `0` (non-walkable, gray) | `1` (walkable floor, light green)
- **Interaction**: 
  - Click to toggle between 0 and 1
  - **Drag to paint/erase**: Hold and drag mouse to batch edit tiles
  - Drag automatically detects whether to set or clear based on first clicked tile
- **Constraints**: Changing to 0 automatically clears static items and land monsters

#### Static Layer
- **Values**: `0` (no items) | `1` (can place items, orange corner marker)
- **Interaction**: Click to toggle, only works on ground tiles (ground = 1)
- **Constraints**: Cannot be set to 1 on non-ground tiles

#### Monster Layer
- **Values**: 
  - `0` (empty)
  - `1` (land monster spawn, pink circle) - ground only
  - `2` (flying monster spawn, blue circle) - anywhere
- **Interaction**: 
  - On ground tiles: `0 → 1 → 2 → 0`
  - On non-ground tiles: `0 → 2 → 0`

## Controls

- **Layer Selection**: Click layer buttons to switch editing mode
- **Visibility Toggle**: Eye icons to show/hide each layer
- **Ground Layer Editing**:
  - **Single Click**: Toggle individual tiles between walkable/non-walkable
  - **Drag Operation**: Hold mouse button and drag to paint/erase multiple tiles
    - First tile determines mode: drag on empty = paint, drag on floor = erase
    - Visual feedback shows drag cursor and blue outline during operation
- **New Template**: Create templates with custom dimensions (1-200)
- **Import**: Load JSON files with automatic validation and correction
- **Export**: Save as single JSON or separate layer files
- **Copy**: Copy JSON to clipboard for easy sharing

## Visual Guide

| Element | Appearance | Description |
|---------|------------|-------------|
| Ground (walkable) | Light green background | Safe to walk on |
| Ground (non-walkable) | Gray background | Cannot walk on |
| Static items | Orange corner "S" | Item placement zones |
| Land monsters | Pink circle with "L" | Ground-only spawns |
| Flying monsters | Blue circle with "F" | Can spawn anywhere |
| Flying (non-ground) | Dark blue circle | Flying monsters on walls |

## Sample Template

A sample 15×11 template is included in `sample-template.json` that demonstrates:
- Basic room layout with walls and walkable areas
- Strategic item placement locations
- Mixed monster spawn types
- Proper constraint adherence

## Development

Built with:
- React + TypeScript
- Vite for development
- Zustand for state management
- DOM-based grid rendering

### Project Structure

```
src/
├── components/           # React components
│   ├── GridEditor.tsx   # Main grid editor
│   ├── Toolbar.tsx      # Top toolbar with controls
│   ├── LayerControl.tsx # Layer visibility and selection
│   ├── ImportExport.tsx # File operations
│   ├── InfoPanel.tsx    # Rules and current state
│   └── NewTemplateDialog.tsx # Template creation
├── store/
│   └── templateStore.ts # Zustand state management
├── types/
│   └── template.ts      # TypeScript interfaces
├── utils/
│   ├── templateUtils.ts # Core logic functions
│   └── fileUtils.ts     # File operations
└── App.tsx              # Main application
```

### Key Functions

- `createEmptyTemplate(width, height)` - Initialize blank template
- `toggleGround(template, x, y)` - Toggle ground state with constraint handling
- `toggleStatic(template, x, y)` - Toggle static items (ground-only)
- `toggleMonster(template, x, y)` - Cycle monster spawns with smart rules
- `validateTemplate(data)` - Check template validity
- `sanitizeTemplate(data)` - Auto-correct invalid templates

## Contributing

This is a frontend-only implementation designed for easy extension. Future enhancements might include:
- Backend integration for template storage
- Advanced editing tools (brush, fill, undo/redo)
- Collaborative editing features
- Template library and sharing