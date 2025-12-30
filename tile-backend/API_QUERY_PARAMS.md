# API Query Parameters Documentation

## GET /api/v1/templates - List Templates with Advanced Filtering

This endpoint now supports comprehensive filtering based on template properties.

### Basic Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `limit` | integer | Number of results to return (1-100, default: 20) | `limit=50` |
| `offset` | integer | Offset for pagination (default: 0) | `offset=20` |
| `name_like` | string | Filter by template name (case-insensitive partial match) | `name_like=boss` |

### Template Type Filter

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `room_type` | string | Filter by room type: `full`, `bridge`, or `platform` | `room_type=bridge` |

### Walkable Ratio Filters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `min_walkable_ratio` | float | Minimum walkable ratio (0.0-1.0) | `min_walkable_ratio=0.5` |
| `max_walkable_ratio` | float | Maximum walkable ratio (0.0-1.0) | `max_walkable_ratio=0.9` |

### Layer Count Filters

#### Static Layer
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `min_static_count` | integer | Minimum number of static tiles | `min_static_count=1` |
| `max_static_count` | integer | Maximum number of static tiles | `max_static_count=10` |

#### Turret Layer
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `min_turret_count` | integer | Minimum number of turret tiles | `min_turret_count=2` |
| `max_turret_count` | integer | Maximum number of turret tiles | `max_turret_count=5` |

#### MobGround Layer
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `min_mobground_count` | integer | Minimum number of mob ground tiles | `min_mobground_count=5` |
| `max_mobground_count` | integer | Maximum number of mob ground tiles | `max_mobground_count=20` |

#### MobAir Layer
| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `min_mobair_count` | integer | Minimum number of mob air tiles | `min_mobair_count=1` |
| `max_mobair_count` | integer | Maximum number of mob air tiles | `max_mobair_count=15` |

### Room Attributes Filters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `has_boss` | boolean | Filter by boss room attribute | `has_boss=true` |
| `has_elite` | boolean | Filter by elite room attribute | `has_elite=true` |
| `has_mob` | boolean | Filter by mob room attribute | `has_mob=true` |
| `has_treasure` | boolean | Filter by treasure room attribute | `has_treasure=true` |
| `has_teleport` | boolean | Filter by teleport room attribute | `has_teleport=true` |
| `has_story` | boolean | Filter by story room attribute | `has_story=true` |

### Door Connectivity Filters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `top_door_connected` | boolean | Filter by top door connectivity | `top_door_connected=true` |
| `right_door_connected` | boolean | Filter by right door connectivity | `right_door_connected=true` |
| `bottom_door_connected` | boolean | Filter by bottom door connectivity | `bottom_door_connected=true` |
| `left_door_connected` | boolean | Filter by left door connectivity | `left_door_connected=true` |

## Example Queries

### Find boss rooms with high walkable ratio
```
GET /api/v1/templates?has_boss=true&min_walkable_ratio=0.7
```

### Find bridge rooms with turrets
```
GET /api/v1/templates?room_type=bridge&min_turret_count=1
```

### Find rooms with all doors connected
```
GET /api/v1/templates?top_door_connected=true&right_door_connected=true&bottom_door_connected=true&left_door_connected=true
```

### Find mob rooms with specific mob counts
```
GET /api/v1/templates?has_mob=true&min_mobground_count=5&max_mobground_count=15
```

### Complex query combining multiple filters
```
GET /api/v1/templates?room_type=platform&min_walkable_ratio=0.4&max_walkable_ratio=0.7&has_elite=true&min_static_count=2&top_door_connected=true&bottom_door_connected=true
```

## Response Format

The response includes all computed fields:

```json
{
  "total": 42,
  "items": [
    {
      "id": "uuid-here",
      "name": "template-name",
      "version": 1,
      "width": 10,
      "height": 8,
      "thumbnail": "base64-string",
      "walkable_ratio": 0.675,
      "room_type": "bridge",
      "room_attributes": {
        "boss": false,
        "elite": true,
        "mob": true,
        "treasure": false,
        "teleport": false,
        "story": false
      },
      "doors_connected": {
        "top": true,
        "right": true,
        "bottom": true,
        "left": false
      },
      "static_count": 3,
      "turret_count": 2,
      "mobground_count": 12,
      "mobair_count": 5,
      "created_at": "2025-01-15T10:30:00Z",
      "updated_at": "2025-01-15T10:30:00Z"
    }
  ]
}
```

## Database Migration

To enable these features, run the migration:

```bash
psql -d tile_templates -f migrations/003_add_computed_fields.up.sql
```

To rollback:

```bash
psql -d tile_templates -f migrations/003_add_computed_fields.down.sql
```

## Notes

- All filters can be combined using AND logic
- Boolean parameters accept: `true`, `false`, `1`, `0`
- Float parameters use dot notation: `0.5`, `0.75`
- Computed fields are automatically calculated when creating templates
- Door connectivity checks if door cells on the ground layer are walkable (value=1)
