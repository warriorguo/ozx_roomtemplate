# 房间难度系数计算方案

> 最后更新：2026-03-20

## 概述

房间生成完成后，自动计算一个综合难度评分（0-1）。评分由**地形难度**和**敌人难度**两个维度加权合成。

```
overall = terrain × 0.4 + enemy × 0.6
```

## 输出格式

生成 API 响应新增 `difficulty` 字段（与 `payload`、`debugInfo` 同级）：

```json
{
  "payload": { ... },
  "debugInfo": { ... },
  "difficulty": {
    "terrain": 0.19,
    "enemy": 0.86,
    "overall": 0.59,
    "details": {
      "groundCoverage": 1.00,
      "narrowPassages": 0,
      "softEdgeCount": 0,
      "pathTortuosity": 1.41,
      "pathStaticBlocks": 12,
      "islandCount": 1,
      "chaserCount": 7,
      "zonerCount": 0,
      "dpsCount": 5,
      "mobAirCount": 4,
      "enemyDensity": 0.0667,
      "enemyConcentration": 0.87
    }
  }
}
```

## 维度一：地形难度 (terrain)

| 因素 | 权重 | 计算方法 | 归一化 |
|------|------|----------|--------|
| Ground 覆盖率 | 25% | walkable / total_area | (1 - coverage) / 0.7，覆盖率 30% 时满分 |
| 窄通道数量 | 20% | 正交邻居 ≤ 2 且为直线通道的 ground 格数 | count / 10，10 个通道时满分 |
| SoftEdge 数量 | 10% | softEdge 层中 value=1 的格子总数 | count / perimeter，perimeter = 2×(w+h) |
| 主路径曲折度 | 20% | mainPath 格子总数 / 路径端点直线距离 | (tortuosity - 1) / 2，ratio=3 时满分 |
| 主路径被 Static 阻挡 | 15% | mainPath 2 格内的 static 格子数 | count / (area × 0.05) |
| 浮岛数量 | 10% | 断开的 ground 区域数（flood fill） | (islands - 1) / 3，4 个岛时满分 |

### 窄通道定义

一个 ground 格子满足以下条件之一时被计为窄通道：
- 恰好 2 个正交邻居且为对向（水平或垂直直线通道）
- 恰好 1 个正交邻居（死胡同）

### 路径曲折度计算

```
tortuosity = path_cell_count / euclidean_distance(path_min, path_max)
```

- tortuosity = 1.0 → 完全直线
- tortuosity = 2.0 → 中等弯曲
- tortuosity ≥ 3.0 → 非常曲折

## 维度二：敌人难度 (enemy)

| 因素 | 权重 | 计算方法 | 归一化 |
|------|------|----------|--------|
| 加权敌人数 | 50% | Chaser×1.5 + Zoner×2.0 + DPS×1.0 + MobAir×0.8 | weighted / (area × 0.08) |
| 敌人密度 | 30% | 敌人总数 / walkable 面积 | density / 0.1 |
| 敌人集中度 | 20% | 1 - (variance / max_variance) | 方差越小 → 集中度越高 |

### 敌人权重说明

| 类型 | 权重 | 理由 |
|------|------|------|
| Chaser | 1.5 | 主动追击，持续压力 |
| Zoner | 2.0 | 区域封锁，限制走位空间 |
| DPS | 1.0 | 定点输出，基础威胁 |
| MobAir | 0.8 | 飞行干扰，间接威胁 |

### 敌人集中度计算

```
centroid = (avg_x, avg_y) of all enemy positions
variance = avg((x - centroid_x)² + (y - centroid_y)²)
max_variance = (width² + height²) / 4
concentration = 1 - clamp01(variance / max_variance)
```

- concentration = 1.0 → 所有敌人扎堆在一起
- concentration = 0.0 → 敌人均匀分布在地图各处

## 难度参考值

| 阶段 | 预期 terrain | 预期 enemy | 预期 overall |
|------|------------|----------|------------|
| Teaching | 0.05-0.20 | 0.10-0.25 | 0.08-0.23 |
| Building | 0.10-0.25 | 0.20-0.40 | 0.16-0.34 |
| Pressure | 0.15-0.35 | 0.50-0.75 | 0.36-0.59 |
| Peak | 0.20-0.40 | 0.70-0.95 | 0.50-0.73 |
| Release | 0.05-0.15 | 0.00-0.15 | 0.02-0.15 |
| Boss | 0.15-0.30 | 0.00-0.10 | 0.06-0.16 |

> 注：以上为 20×12 房间的参考范围，实际值受随机种子影响。

## 相关代码

- `tile-backend/internal/generate/difficulty.go` — 计算逻辑
- 所有生成器的响应中包含 `difficulty` 字段
