# 文档与实现差异记录

本文档记录 bridge 和 platform 房间生成规则中，**文档描述**与**实际代码实现**之间的差异。

> 最后更新：2026-03-19

## 差异总览

| # | 文件 | 位置 | 严重程度 | 状态 |
|---|------|------|----------|------|
| 1 | bridge.go | 浮岛继续概率 | 🔴 高 | 待修正 |
| 2 | bridge.go | 平台笔刷尺寸 | 🔴 高 | 待修正 |
| 3 | bridge.go | 连接笔刷尺寸 | 🟡 中 | 待修正 |
| 4 | bridge.go | 浮岛距离约束描述 | 🟢 低 | 待修正 |

---

## 差异 1：浮岛继续放置概率

**文件**: `internal/generate/bridge.go:828`

**文档描述** (`bridge-generation-rules.md` Step 3.2):
> 50% probability check: With 50% probability, skip this area

**实际代码**:
```go
// Step 2: 50% chance to continue
if rand.Float64() >= 0.8 {
```

**分析**:
- 代码中 `rand.Float64() >= 0.8` 意味着只有 **20%** 概率停止（即 **80%** 概率继续放置）
- 文档描述的是 **50%** 概率停止
- 注释写的是 "50% chance to continue"，但阈值是 0.8 而非 0.5

**影响**: 实际生成的浮岛数量**远多于**文档预期。这会显著影响 bridge 房间的视觉密度和可玩性。

**建议**: 确认设计意图后，统一文档和代码。如果 80% 是有意调整的，则更新文档和注释；如果应该是 50%，则修改阈值为 0.5。

---

## 差异 2：Bridge 平台笔刷尺寸

**文件**: `internal/generate/bridge.go:204-208`

**文档描述** (`bridge-generation-rules.md` Step 2.3):
> Randomly select a brush from: **2×2**, **2×3**, **3×3**, **3×2**, **4×3**, **3×4**, **4×4**, **4×5**, **5×4**, **5×5**

**实际代码**:
```go
var platformBrushes = []BrushSize{
	/*{2, 2}, {2, 3}, {3, 3}, {3, 2},
	{4, 3}, {3, 4}, {4, 4}, {4, 5},
	{5, 4}, {5, 5},*/
	{4, 4}, {6, 6},
}
```

**分析**:
- 原有的 10 种笔刷尺寸被全部注释掉
- 当前仅使用 **4×4** 和 **6×6** 两种尺寸
- 缺少小尺寸（2×2, 2×3, 3×2）和中等尺寸（3×3, 4×3, 3×4 等）

**影响**:
- 生成的平台只有两种固定大小，**变化性大幅降低**
- 平台整体偏大（最小 4×4），小房间中可能覆盖过多区域
- 文档中描述的细粒度尺寸选择不再存在

**建议**: 更新文档以反映当前仅使用 4×4 和 6×6 的实际情况，或根据设计需要恢复部分中间尺寸。

---

## 差异 3：门连接笔刷尺寸

**文件**: `internal/generate/bridge.go:200-202`

**文档描述** (`bridge-generation-rules.md` Step 1):
> Randomly select a brush size from: **2×2**, **3×3**, **4×4**

**实际代码**:
```go
var connectionBrushes = []BrushSize{
	{2, 2} /*{3, 3},*/, {4, 4},
}
```

**分析**:
- **3×3** 笔刷被注释掉
- 当前仅使用 **2×2** 和 **4×4**

**影响**: 缺少中等宽度的连接路径。路径要么窄（2 格）要么宽（4 格），没有 3 格的中间选项。视觉上可能产生较大的尺寸跳跃感。

**建议**: 更新文档移除 3×3 的描述，或根据设计需要恢复 3×3 笔刷。

---

## 差异 4：浮岛距离约束描述不够精确

**文件**: `internal/generate/bridge.go:737, 909-965`

**文档描述** (`bridge-generation-rules.md` Step 3.2):
> Distance constraint: Island must be exactly **2 cells away** from existing ground (not closer, not farther)

**实际代码**（双层约束）:
```go
const minIslandGroundDistance = 2

// 第一层：岛屿周围 minIslandGroundDistance(2) 格内不能有地面
checkStartX := x - minIslandGroundDistance
// ...
// The margin area must be void (no existing ground within minIslandGroundDistance)

// 第二层：距离 minIslandGroundDistance+1(3) 格处必须存在地面
outerDist := minIslandGroundDistance + 1
// Additionally, check that there IS ground just outside the margin (at distance exactly minIslandGroundDistance+1)
```

**分析**:
- 文档简化为"恰好 2 格"
- 实际实现是两个独立约束：
  1. 2 格内**不能有**地面（内圈 margin 检查）
  2. 3 格处**必须有**地面（外圈 ring 检查）
- 文档的 Step 3.3 "Distance Constraint Details" 部分对此描述更准确，但 Step 3.2 的简化描述有误导

**影响**: 较小，主要是文档表述不够精确。功能行为本身是合理的。

**建议**: 将 Step 3.2 的描述更新为："Island must have no ground within 2 cells AND must have ground at exactly 3 cells distance"，与 Step 3.3 保持一致。

---

## Platform 生成规则一致性

Platform 房间的文档与实现**基本一致**，以下已验证匹配：

| 规则 | 状态 |
|------|------|
| Strategy 选择（50/50） | ✅ 一致 |
| 橡皮擦操作次数（0-3） | ✅ 一致 |
| 橡皮擦方法列表 | ✅ 一致 |
| 角落橡皮擦概率分布 | ✅ 一致 |
| 连通性回滚机制 | ✅ 一致 |
| 尺寸范围（10-200） | ✅ 一致 |

---

## 共享层生成规则一致性

以下共享层（bridge 和 platform 通用）已验证文档与实现一致：

| 层 | 状态 |
|----|------|
| Soft Edge | ✅ 一致 |
| Bridge Layer | ✅ 一致 |
| Static | ✅ 一致 |
| Chaser | ✅ 一致（替代旧 Turret 层） |
| Zoner | ✅ 一致（替代旧 MobGround 层） |
| DPS | ✅ 一致（新增层） |
| MobAir | ✅ 一致 |

> **注意**：2026-03-19 起，旧的 turret/mobGround 系统已被新的 chaser/zoner/dps 敌人系统替代。
> 旧的 Turret 层和 MobGround 层不再存在，相关的一致性验证已不再适用。
> 新增的 mainPath 和 stageType 字段也已加入 API 响应。
> 详细的敌人系统规则请参考 [enemy-system-rules.md](enemy-system-rules.md)。

---

## 修正历史

| 日期 | 差异 # | 操作 | 说明 |
|------|--------|------|------|
| 2026-03-19 | N/A | 敌人系统替换 | 旧 turret/mobGround 系统被 chaser/zoner/dps 系统替代，文档已同步更新 |
