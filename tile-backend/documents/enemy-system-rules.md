# Enemy System & Stage Rules

本文档描述新的敌人系统和阶段规则。替代了旧的 Turret/MobGround 系统。

> 最后更新：2026-03-19

## 敌人类型

| 类型 | 角色 | 地面要求 | 主路径距离 | 放置偏好 |
|------|------|----------|-----------|----------|
| Chaser | 压迫 | ground=1 | 0-3 | 低脆皮攻击度（容易到达） |
| Zoner | 控场 | ground=1 | 0-5 | 高脆皮攻击度（远程优势） |
| DPS | 输出 | ground=1 | 0-4 | 靠近 Chaser/Static |
| MobAir | 干扰 | 不需要 | - | Zoner/Chaser 密集区 |

## 主路径 (Main Path)

### 计算方法

在 bridge 层生成后，使用中心偏向的 Dijkstra 算法寻找连通所有门的路径：

1. 构建可行走网格（ground + bridge）
2. 对每对门，找到经过房间中心的最短路径（靠近中心的格子代价更低）
3. 合并所有门-门路径为主路径集合

### 距离度量

对地图上任意一格子 C，设最近主路径格子为 R：

| 度量 | 定义 |
|------|------|
| 主路径直线距离 | C 和 R 之间的曼哈顿距离 |
| 主路径步行距离 | C 到 R 的 BFS 步行距离（仅走可行走格子） |
| 脆皮攻击度 | 步行距离 / 直线距离（值越高越适合远程攻击） |

### 代价函数

```
cellCost(p) = 1.0 + 2.0 × (distToCenter / maxDist)
```

靠近中心的格子代价 ≈ 1.0，远离中心的格子代价最高 ≈ 3.0。

## 门禁区

统一规则：门位置曼哈顿距离 ≤ 2 的所有格子为禁区，禁止放置任何实体。

```go
const doorForbiddenRadius = 2
```

## 按种类的部署策略

### Zoner（控场）

- 主路径距离：0-5
- 优先选择**脆皮攻击度高**的点
- Zoner 与主路径之间不能存在 static（Bresenham 直线检查）
- 不能重叠：softEdge, bridge, rail, static
- 相邻约束：不能与其他 Zoner 8 方向相邻

### Chaser（压迫）

- 主路径距离：0-3
- 优先选择**脆皮攻击度低**的点（容易被玩家到达）
- 不能重叠：softEdge, bridge, rail, static, zoner
- 相邻约束：不能与其他 Chaser 8 方向相邻

### DPS（输出）

- 主路径距离：0-4
- 可以在 Chaser/Static 附近（放宽约束，支援角色）
- 偏好评分：靠近 Chaser +3分，靠近 Static +2分，加上脆皮攻击度 ×0.5
- 不能重叠：bridge, rail, zoner
- 相邻约束：不能与其他 DPS 8 方向相邻

### MobAir（干扰）

- 不需要 ground
- 优先选择 Zoner/Chaser **密集**的区域
- 门距离 ≥ 4（曼哈顿距离）
- 边缘距离 ≥ 2
- 相邻约束：不能与其他 MobAir 8 方向相邻（距离 ≥ 1）

## 生成顺序

```
Zoner → Chaser → DPS → MobAir
```

Zoner 先放置（占据高脆皮攻击度位置），然后 Chaser 填充近路径位置，DPS 在 Chaser/Static 附近，最后 MobAir 在密集区。

## 阶段规则 (Stage Rules)

### 阶段类型

| 阶段 | 英文 | 说明 |
|------|------|------|
| 引导期 | teaching | 建立控制感 |
| 建立期 | building | 引入差异 |
| 压力期 | pressure | 组合出现 |
| 峰值期 | peak | near miss |
| 释放期 | release | 放礼物等 |
| Boss | boss | Boss 战 |

### 阶段配置

#### 引导期 (teaching)
- DPS：2-3
- 其他：0

#### 建立期 (building)
- DPS：2-3
- Chaser：2-3
- 其他：0

#### 压力期 (pressure)
- **房间限制**：不能是 bridge
- DPS：4-6
- Chaser：6-8
- Zoner：1
- MobAir：2-4

#### 峰值期 (peak)
- **房间限制**：只能是 full
- **门限制**：不能是 2 开门的对角组合（左上/左下/右上/右下）
- DPS：6-12
- Chaser：6-8
- Zoner：2-3
- MobAir：2-4

#### 释放期 (release)
- DPS：0-3
- 其他：0

#### Boss 期 (boss)
- **房间限制**：不能是 bridge
- **门限制**：只能是对角 2 门组合或 1 门
- **场地要求**：中间必须存在 6×6 空地，距离边缘 > 3
- 无敌人（由 Boss 机制决定）

### 阶段验证流程

1. 检查 roomType 是否在 allowedRoomTypes 中
2. 检查门配置是否满足 DoorRestrictions
3. 随机生成范围内的敌人数量
4. Boss 阶段额外检查 6×6 空地

## API 参数

### 请求

```json
{
  "width": 20,
  "height": 12,
  "doors": ["top", "bottom", "left", "right"],
  "softEdgeCount": 2,
  "railEnabled": true,
  "staticCount": 4,
  "chaserCount": 3,
  "zonerCount": 2,
  "dpsCount": 4,
  "mobAirCount": 3,
  "stageType": "pressure"
}
```

当指定 `stageType` 时，`chaserCount`/`zonerCount`/`dpsCount`/`mobAirCount` 会被阶段规则覆盖。

### 响应新增字段

```json
{
  "payload": {
    "chaser": [[0,1,...], ...],
    "zoner": [[0,1,...], ...],
    "dps": [[0,1,...], ...],
    "mainPath": [[0,1,...], ...],
    "stageType": "pressure"
  },
  "debugInfo": {
    "mainPath": {
      "pathCellCount": 31,
      "pathSegments": ["(10,0)->(10,11) len=12", ...]
    },
    "chaser": { "targetCount": 3, "placedCount": 3, ... },
    "zoner": { "targetCount": 2, "placedCount": 2, ... },
    "dps": { "targetCount": 4, "placedCount": 4, ... }
  }
}
```

## 常量

```go
const doorForbiddenRadius = 2
const chaserMaxPathDist   = 3
const zonerMaxPathDist    = 5
const dpsMaxPathDist      = 4
const mobAirMinDoorDistance = 4
const mobAirMinEdgeDistance = 2
```
