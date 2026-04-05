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

Bulk-download all tilemaps from the backend API and save them locally, organized by
`roomCategory` subfolder with the standard `{roomShape}_{stageType}_{seq}.json` naming.

## Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `endpoint` | string | `https://ozx-roomtpl.local.playquota.com/api/v1` | Backend API base URL |
| `targetDir` | string | *(ask user)* | Local directory to save tilemaps (e.g. `Assets/StreamingAssets/TilemapData`) |
| `roomCategory` | string | *(all)* | Filter by category (`normal`, `basement`, `test`, `cave`), or leave empty to sync all |
| `overwrite` | bool | `false` | Whether to overwrite existing files. If false, skip files that already exist. |

## Workflow

### Step 1: Confirm parameters

Display parameters and ask the user to confirm or adjust:

```
Room Sync Parameters:
  endpoint:      https://ozx-roomtpl.local.playquota.com/api/v1
  targetDir:     (not set - please provide)
  roomCategory:  (all)
  overwrite:     false
```

The `targetDir` is required. If not provided, ask the user.

### Step 2: Fetch all templates

Paginate through the templates list API:

```bash
curl -sL "https://ozx-roomtpl.local.playquota.com/api/v1/templates?limit=100&offset=0"
```

Response: `{"total": N, "items": [...]}`

Continue fetching with increasing `offset` until all templates are retrieved.
Each item in `items` has an `id` field — collect all IDs.

If `roomCategory` filter is set, skip items where `room_category` doesn't match
(note: the list API uses `room_category` with underscore in the summary response,
but the payload uses `roomCategory` camelCase).

### Step 3: Download each template payload

For each template ID:

```bash
curl -sL "https://ozx-roomtpl.local.playquota.com/api/v1/templates/{id}"
```

Extract the `payload` object from the response. This contains:
- `roomShape`: `"all"`, `"bridge"`, `"platform"`, or null
- `roomCategory`: `"normal"`, `"basement"`, `"test"`, `"cave"`, or null
- `stageType`: `"default"`, `"teaching"`, `"building"`, etc., or null
- All layer data (ground, static, chaser, etc.)

### Step 4: Save with naming convention

For each downloaded payload:

1. **Determine folder**: `{targetDir}/{roomCategory}/`
   - Use `roomCategory` from the payload, default to `"normal"` if null/empty
   - Create the subfolder if it doesn't exist

2. **Determine filename**: `{roomShape}_{stageType}_{seq}.json`
   - `roomShape`: from payload, or `"none"` if null
   - `stageType`: from payload, or `"default"` if null/empty
   - `seq`: two-digit number, auto-incremented per `{roomShape}_{stageType}_` prefix

3. **Auto-increment logic**:
   - Use Glob to find `{targetDir}/{roomCategory}/{roomShape}_{stageType}_*.json`
   - Extract the highest existing sequence number
   - Next file gets highest + 1 (or `01` if none exist)
   - Track sequences in memory during the sync to avoid re-scanning

4. **Check overwrite**: If `overwrite` is false and file exists, skip it

5. **Save payload only** — write the payload JSON (not the template wrapper with id/name/timestamps).
   Use `python3 -c` or the Write tool.

### Step 5: Report summary

After all templates are processed, display a summary:

```
Sync complete!
  Total templates: 25
  Saved: 22
  Skipped (already exists): 3
  Errors: 0

  By category:
    normal/    — 18 files (12 all, 4 bridge, 2 platform)
    basement/  — 4 files (2 all, 2 bridge)
    cave/      — 3 files (3 all)
```

## Batch Download Script

For efficiency, use a single python script to handle the full sync rather than
individual curl calls per template. Here's the approach:

```bash
python3 -c "
import json, subprocess, os, glob, re

endpoint = '{endpoint}'
target = '{targetDir}'
overwrite = {overwrite}
category_filter = '{roomCategory}' or None

# 1. Fetch all template IDs
offset, all_ids = 0, []
while True:
    out = subprocess.check_output(['curl', '-sL', f'{endpoint}/templates?limit=100&offset={offset}'])
    data = json.loads(out)
    items = data.get('items') or []
    for item in items:
        if category_filter and item.get('room_category') != category_filter:
            continue
        all_ids.append(item['id'])
    if offset + len(items) >= data['total']:
        break
    offset += 100

print(f'Found {len(all_ids)} templates to sync')

# 2. Download and save each
saved, skipped, errors = 0, 0, 0
seq_counters = {}  # (category, shape, stage) -> next seq

for tid in all_ids:
    try:
        out = subprocess.check_output(['curl', '-sL', f'{endpoint}/templates/{tid}'])
        tpl = json.loads(out)
        payload = tpl['payload']

        cat = payload.get('roomCategory') or 'normal'
        shape = payload.get('roomShape') or 'none'
        stage = payload.get('stageType') or 'default'

        folder = os.path.join(target, cat)
        os.makedirs(folder, exist_ok=True)

        # Auto-increment sequence
        key = (cat, shape, stage)
        if key not in seq_counters:
            existing = glob.glob(os.path.join(folder, f'{shape}_{stage}_*.json'))
            nums = [int(re.search(r'_(\d+)\.json$', f).group(1)) for f in existing if re.search(r'_(\d+)\.json$', f)]
            seq_counters[key] = max(nums) + 1 if nums else 1

        seq = seq_counters[key]
        seq_counters[key] = seq + 1
        filename = f'{shape}_{stage}_{seq:02d}.json'
        filepath = os.path.join(folder, filename)

        if not overwrite and os.path.exists(filepath):
            skipped += 1
            continue

        with open(filepath, 'w') as f:
            json.dump(payload, f, separators=(',', ':'))
        saved += 1
        print(f'  Saved: {cat}/{filename}')

    except Exception as e:
        errors += 1
        print(f'  Error on {tid}: {e}')

print(f'\nDone: {saved} saved, {skipped} skipped, {errors} errors')
"
```

Replace `{endpoint}`, `{targetDir}`, `{overwrite}`, `{roomCategory}` with actual values.

## Error Handling

- **Backend not reachable**: Show connection error and suggest checking the endpoint URL
- **Empty template list**: Report "0 templates found" — not an error
- **Invalid payload** (missing roomShape/stageType): Use defaults (`"none"`, `"default"`)
- **File write permission error**: Report which file failed and continue with remaining
