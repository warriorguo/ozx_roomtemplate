---
name: room-sync
description: >
  Bulk-download all tilemaps from the OZX Room Template backend and save to local
  TilemapData directory following the standard folder and naming conventions.
  Use this skill when the user wants to sync tilemaps, download all room templates,
  export tilemaps to Unity, pull tilemap data, or refresh local tilemap files.
  Triggers on: "sync tilemaps", "download all tilemaps", "pull tilemaps",
  "export tilemaps to unity", "refresh tilemap data", "sync rooms".
---

# Room Sync

Bulk-download all tilemaps from the backend API and save locally.

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `endpoint` | string | `https://ozx-roomtpl.local.playquota.com/api/v1` | Backend API base URL |
| `targetDir` | string | *(ask user)* | Local directory to save tilemaps |
| `roomCategory` | string | *(all)* | Filter by category, or empty for all |
| `overwrite` | bool | `false` | Whether to overwrite existing files |

## Usage

1. Ask the user for `targetDir` if not provided
2. Substitute the 4 parameters into the script below
3. Run it via Bash tool

```python
import json, subprocess, os, glob, re
from collections import Counter

# --- Door remapping for OZX Unity (transpose: top↔left, bottom↔right) ---
DOOR_REMAP = {'top': 'left', 'right': 'bottom', 'bottom': 'right', 'left': 'top'}
BITMASK_REMAP = {1: 8, 2: 4, 4: 2, 8: 1}  # Top=1→Left=8, Right=2→Bottom=4, etc.

def remap_doors(payload):
    """Remap door directions for OZX Unity coordinate convention."""
    if 'doors' in payload and payload['doors']:
        old = payload['doors']
        payload['doors'] = {DOOR_REMAP[k]: v for k, v in old.items() if k in DOOR_REMAP}
    if 'openDoors' in payload and payload['openDoors'] is not None:
        old_mask = payload['openDoors']
        new_mask = 0
        for bit, remapped_bit in BITMASK_REMAP.items():
            if old_mask & bit:
                new_mask |= remapped_bit
        payload['openDoors'] = new_mask
    return payload

ENDPOINT = '{endpoint}'
TARGET = '{targetDir}'
OVERWRITE = {overwrite}  # True or False
CATEGORY_FILTER = '{roomCategory}' or None  # None = all categories

# --- Fetch all template IDs ---
offset, all_ids = 0, []
while True:
    out = subprocess.check_output(
        ['curl', '-sL', f'{ENDPOINT}/templates?limit=100&offset={offset}'])
    data = json.loads(out)
    items = data.get('items') or []
    for item in items:
        if CATEGORY_FILTER and item.get('room_category') != CATEGORY_FILTER:
            continue
        all_ids.append((item['id'], item.get('name', '')))
    if offset + len(items) >= data.get('total', 0):
        break
    offset += 100

print(f'Found {len(all_ids)} templates to sync')
if not all_ids:
    raise SystemExit(0)

# --- Download and save each ---
saved, skipped, errors = 0, 0, 0
seq_counters = {}
by_category = {}

for tid, tname in all_ids:
    try:
        out = subprocess.check_output(
            ['curl', '-sL', f'{ENDPOINT}/templates/{tid}'])
        tpl = json.loads(out)
        payload = tpl['payload']

        remap_doors(payload)

        cat = payload.get('roomCategory') or 'normal'
        shape = payload.get('roomShape') or 'none'
        stage = payload.get('stageType') or 'default'
        doors = payload.get('openDoors', 0)

        folder = os.path.join(TARGET, cat)
        os.makedirs(folder, exist_ok=True)

        # Auto-increment sequence per (category, shape, stage, doors)
        key = (cat, shape, stage, doors)
        if key not in seq_counters:
            pat = os.path.join(folder, f'{shape}_{stage}_{doors}_*.json')
            nums = [int(m.group(1))
                    for f in glob.glob(pat)
                    if (m := re.search(r'_(\d+)\.json$', f))]
            seq_counters[key] = max(nums) + 1 if nums else 1

        seq = seq_counters[key]
        seq_counters[key] = seq + 1
        filename = f'{shape}_{stage}_{doors}_{seq:02d}.json'
        filepath = os.path.join(folder, filename)

        if not OVERWRITE and os.path.exists(filepath):
            skipped += 1
            continue

        with open(filepath, 'w') as f:
            json.dump(payload, f, separators=(',', ':'))
        saved += 1
        by_category.setdefault(cat, []).append(shape)
        print(f'  Saved: {cat}/{filename}')

    except Exception as e:
        errors += 1
        print(f'  Error on {tid}: {e}')

# --- Summary ---
print(f'\nSync complete!')
print(f'  Total: {len(all_ids)}  Saved: {saved}  Skipped: {skipped}  Errors: {errors}')
for cat, shapes in sorted(by_category.items()):
    detail = ', '.join(f'{v} {k}' for k, v in Counter(shapes).most_common())
    print(f'  {cat}/ -- {len(shapes)} files ({detail})')
```

### Naming convention

```
{targetDir}/{roomCategory}/{roomShape}_{stageType}_{openDoors}_{seq}.json
```

- `roomCategory`: subfolder — `normal`, `basement`, `test`, `cave` (default: `normal`)
- `roomShape`: `all`, `bridge`, `platform` (default: `none`)
- `stageType`: `default`, `start`, `teaching`, `building`, `pressure`, `peak`, `release`, `boss`
- `openDoors`: bitmask int — Top=1, Right=2, Bottom=4, Left=8 (e.g. 15 = all doors)
- `seq`: two-digit auto-increment per prefix
