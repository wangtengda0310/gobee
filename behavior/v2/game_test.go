package behavior

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 游戏业务场景测试
// =============================================================================

// TestGameAI_PatrolAndChase 测试 NPC 巡逻和追击行为
// 场景：NPC 在巡逻时发现玩家，切换到追击状态
func TestGameAI_PatrolAndChase(t *testing.T) {
	ctx := make(Context)

	// 游戏状态
	ctx["playerVisible"] = false
	ctx["playerDistance"] = 100.0
	ctx["npcState"] = "idle"

	// 巡逻动作
	patrol := NewAction(func(ctx Context) Result {
		ctx["npcState"] = "patroling"
		return Success
	})

	// 检测玩家条件
	canSeePlayer := NewCondition(func(ctx Context) bool {
		return ctx["playerVisible"].(bool)
	})

	// 追击动作
	chase := NewAction(func(ctx Context) Result {
		distance := ctx["playerDistance"].(float64)
		if distance > 10 {
			ctx["playerDistance"] = distance - 5 // 接近玩家
			ctx["npcState"] = "chasing"
			return Running
		}
		ctx["npcState"] = "attacking"
		return Success
	})

	// 构建行为树：如果看到玩家就追击，否则巡逻
	aiTree := NewSelector(
		NewSequence(canSeePlayer, chase),
		patrol,
	)

	// 场景1：没有看到玩家，执行巡逻
	result := aiTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "patroling", ctx["npcState"])

	// 场景2：看到玩家，开始追击
	ctx["playerVisible"] = true
	ctx["playerDistance"] = 50.0
	result = aiTree.Tick(ctx)
	assert.Equal(t, Running, result)
	assert.Equal(t, "chasing", ctx["npcState"])

	// 继续追击直到接近
	for ctx["playerDistance"].(float64) > 10 {
		result = aiTree.Tick(ctx)
		assert.Equal(t, Running, result)
	}

	// 追击完成，可以攻击
	result = aiTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "attacking", ctx["npcState"])
}

// TestGameAI_CombatWithHealthCheck 测试战斗中的血量判断
// 场景：NPC 在战斗中根据血量决定攻击或逃跑
func TestGameAI_CombatWithHealthCheck(t *testing.T) {
	ctx := make(Context)

	// 战斗状态
	ctx["health"] = 100
	ctx["maxHealth"] = 100
	ctx["enemyInView"] = true
	ctx["action"] = ""

	// 条件：血量低于30%
	isLowHealth := NewCondition(func(ctx Context) bool {
		health := ctx["health"].(int)
		maxHealth := ctx["maxHealth"].(int)
		return float64(health)/float64(maxHealth) < 0.3
	})

	// 条件：敌人在视野内
	hasEnemy := NewCondition(func(ctx Context) bool {
		return ctx["enemyInView"].(bool)
	})

	// 逃跑动作
	flee := NewAction(func(ctx Context) Result {
		ctx["action"] = "fleeing"
		return Success
	})

	// 攻击动作
	attack := NewAction(func(ctx Context) Result {
		ctx["action"] = "attacking"
		ctx["health"] = ctx["health"].(int) - 10 // 攻击消耗体力
		return Success
	})

	// 待机动作
	idle := NewAction(func(ctx Context) Result {
		ctx["action"] = "idle"
		return Success
	})

	// 构建行为树：
	// - 如果血量低，逃跑
	// - 如果有敌人且血量正常，攻击
	// - 否则待机
	combatTree := NewSelector(
		NewSequence(isLowHealth, flee),
		NewSequence(hasEnemy, NewInverter(isLowHealth), attack),
		idle,
	)

	// 场景1：满血有敌人，攻击
	result := combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "attacking", ctx["action"])

	// 场景2：血量降低但还不低，继续攻击
	ctx["health"] = 40
	ctx["action"] = ""
	result = combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "attacking", ctx["action"])

	// 场景3：血量低于30%，逃跑
	ctx["health"] = 25
	ctx["action"] = ""
	result = combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "fleeing", ctx["action"])

	// 场景4：没有敌人，待机
	ctx["enemyInView"] = false
	ctx["health"] = 50
	ctx["action"] = ""
	result = combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "idle", ctx["action"])
}

// TestGameAI_SkillCombo 测试技能连招系统
// 场景：按顺序释放技能1 -> 技能2 -> 技能3
func TestGameAI_SkillCombo(t *testing.T) {
	ctx := make(Context)

	// 技能状态
	ctx["comboCount"] = 0
	ctx["skillsUsed"] = []string{}

	// 初始化冷却
	ctx["cooldowns"] = map[string]bool{
		"skill1": true,
		"skill2": true,
		"skill3": true,
	}

	// 技能冷却检查
	skillReady := func(skillName string) Node {
		return NewCondition(func(ctx Context) bool {
			cooldowns := ctx["cooldowns"].(map[string]bool)
			return cooldowns[skillName]
		})
	}

	// 使用技能动作
	useSkill := func(skillName string) Node {
		return NewAction(func(ctx Context) Result {
			skills := ctx["skillsUsed"].([]string)
			ctx["skillsUsed"] = append(skills, skillName)
			ctx["comboCount"] = ctx["comboCount"].(int) + 1
			fmt.Printf("释放技能: %s\n", skillName)
			return Success
		})
	}

	// 构建连招行为树：技能1 -> 技能2 -> 技能3
	comboTree := NewSequence(
		NewSequence(skillReady("skill1"), useSkill("skill1")),
		NewSequence(skillReady("skill2"), useSkill("skill2")),
		NewSequence(skillReady("skill3"), useSkill("skill3")),
	)

	// 执行连招
	result := comboTree.Tick(ctx)
	assert.Equal(t, Success, result)

	skillsUsed := ctx["skillsUsed"].([]string)
	assert.Equal(t, []string{"skill1", "skill2", "skill3"}, skillsUsed)
	assert.Equal(t, 3, ctx["comboCount"])
}

// TestGameAI_SkillComboWithFailure 测试连招中断
// 场景：技能2冷却中，连招中断
func TestGameAI_SkillComboWithFailure(t *testing.T) {
	ctx := make(Context)

	ctx["skillsUsed"] = []string{}
	ctx["cooldowns"] = map[string]bool{
		"skill1": true,
		"skill2": false, // 技能2冷却中
		"skill3": true,
	}

	skillReady := func(skillName string) Node {
		return NewCondition(func(ctx Context) bool {
			cooldowns := ctx["cooldowns"].(map[string]bool)
			return cooldowns[skillName]
		})
	}

	useSkill := func(skillName string) Node {
		return NewAction(func(ctx Context) Result {
			skills := ctx["skillsUsed"].([]string)
			ctx["skillsUsed"] = append(skills, skillName)
			return Success
		})
	}

	comboTree := NewSequence(
		NewSequence(skillReady("skill1"), useSkill("skill1")),
		NewSequence(skillReady("skill2"), useSkill("skill2")),
		NewSequence(skillReady("skill3"), useSkill("skill3")),
	)

	// 执行连招，应该在技能2处中断
	result := comboTree.Tick(ctx)
	assert.Equal(t, Failure, result)

	skillsUsed := ctx["skillsUsed"].([]string)
	assert.Equal(t, []string{"skill1"}, skillsUsed) // 只有技能1被执行
}

// TestGameAI_ResourceManagement 测试资源管理
// 场景：法师施法需要消耗蓝量，蓝量不足时使用药水
func TestGameAI_ResourceManagement(t *testing.T) {
	ctx := make(Context)

	ctx["mana"] = 50
	ctx["maxMana"] = 100
	ctx["potions"] = 3
	ctx["action"] = ""

	// 条件：蓝量足够施法
	hasMana := NewCondition(func(ctx Context) bool {
		return ctx["mana"].(int) >= 30
	})

	// 条件：有药水
	hasPotion := NewCondition(func(ctx Context) bool {
		return ctx["potions"].(int) > 0
	})

	// 施法动作
	castSpell := NewAction(func(ctx Context) Result {
		ctx["mana"] = ctx["mana"].(int) - 30
		ctx["action"] = "casting"
		return Success
	})

	// 使用药水动作
	usePotion := NewAction(func(ctx Context) Result {
		potions := ctx["potions"].(int)
		ctx["potions"] = potions - 1
		ctx["mana"] = ctx["mana"].(int) + 40
		if ctx["mana"].(int) > ctx["maxMana"].(int) {
			ctx["mana"] = ctx["maxMana"]
		}
		ctx["action"] = "using_potion"
		return Success
	})

	// 普通攻击
	normalAttack := NewAction(func(ctx Context) Result {
		ctx["action"] = "normal_attack"
		return Success
	})

	// 构建行为树：
	// - 如果有蓝，施法
	// - 如果没蓝但有药水，使用药水
	// - 否则普通攻击
	mageAI := NewSelector(
		NewSequence(hasMana, castSpell),
		NewSequence(NewInverter(hasMana), hasPotion, usePotion),
		normalAttack,
	)

	// 场景1：有蓝量，施法
	result := mageAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "casting", ctx["action"])
	assert.Equal(t, 20, ctx["mana"])

	// 场景2：蓝量不足，使用药水
	ctx["action"] = ""
	result = mageAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "using_potion", ctx["action"])
	assert.Equal(t, 60, ctx["mana"])
	assert.Equal(t, 2, ctx["potions"])

	// 场景3：再次有蓝量，施法
	ctx["action"] = ""
	result = mageAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "casting", ctx["action"])
}

// TestGameAI_ParallelAttacks 测试并行攻击
// 场景：同时进行普通攻击和召唤宠物攻击
func TestGameAI_ParallelAttacks(t *testing.T) {
	ctx := make(Context)

	ctx["playerAttackReady"] = true
	ctx["petAttackReady"] = true
	ctx["attacksExecuted"] = []string{}

	// 玩家攻击
	playerAttack := NewAction(func(ctx Context) Result {
		attacks := ctx["attacksExecuted"].([]string)
		ctx["attacksExecuted"] = append(attacks, "player")
		return Success
	})

	// 宠物攻击
	petAttack := NewAction(func(ctx Context) Result {
		attacks := ctx["attacksExecuted"].([]string)
		ctx["attacksExecuted"] = append(attacks, "pet")
		return Success
	})

	// 并行执行两个攻击
	// 需要2个成功，1个失败就整体失败
	parallelAttack := NewParallel(2, 1, playerAttack, petAttack)

	result := parallelAttack.Tick(ctx)
	assert.Equal(t, Success, result)

	attacks := ctx["attacksExecuted"].([]string)
	assert.Len(t, attacks, 2)
	assert.Contains(t, attacks, "player")
	assert.Contains(t, attacks, "pet")
}

// TestGameAI_BossPhases 测试 Boss 阶段转换
// 场景：Boss 根据血量进入不同阶段
func TestGameAI_BossPhases(t *testing.T) {
	ctx := make(Context)

	ctx["bossHealth"] = 100
	ctx["bossMaxHealth"] = 100
	ctx["phase"] = 1
	ctx["action"] = ""

	// 阶段1攻击
	phase1Attack := NewAction(func(ctx Context) Result {
		ctx["action"] = "phase1_attack"
		return Success
	})

	// 阶段2攻击（血量低于70%）
	phase2Attack := NewAction(func(ctx Context) Result {
		ctx["action"] = "phase2_attack"
		return Success
	})

	// 阶段3攻击（血量低于30%）
	phase3Attack := NewAction(func(ctx Context) Result {
		ctx["action"] = "phase3_attack"
		return Success
	})

	// 血量检查
	healthBelow := func(percentage float64) Node {
		return NewCondition(func(ctx Context) bool {
			health := ctx["bossHealth"].(int)
			maxHealth := ctx["bossMaxHealth"].(int)
			return float64(health)/float64(maxHealth) < percentage
		})
	}

	// Boss 行为树：根据血量选择阶段
	bossAI := NewSelector(
		NewSequence(healthBelow(0.3), phase3Attack),
		NewSequence(healthBelow(0.7), phase2Attack),
		phase1Attack,
	)

	// 阶段1：满血
	result := bossAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "phase1_attack", ctx["action"])

	// 阶段2：血量低于70%
	ctx["bossHealth"] = 60
	ctx["action"] = ""
	result = bossAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "phase2_attack", ctx["action"])

	// 阶段3：血量低于30%
	ctx["bossHealth"] = 20
	ctx["action"] = ""
	result = bossAI.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "phase3_attack", ctx["action"])
}

// TestGameAI_CooldownSystem 测试冷却系统
// 场景：技能有冷却时间，冷却期间无法使用
func TestGameAI_CooldownSystem(t *testing.T) {
	ctx := make(Context)

	ctx["skillCooldown"] = 0
	ctx["currentTick"] = 0
	ctx["skillUsed"] = false

	// 检查冷却
	cooldownReady := NewCondition(func(ctx Context) bool {
		cooldown := ctx["skillCooldown"].(int)
		currentTick := ctx["currentTick"].(int)
		return currentTick >= cooldown
	})

	// 使用技能并设置冷却
	useSkillWithCooldown := NewAction(func(ctx Context) Result {
		ctx["skillUsed"] = true
		ctx["skillCooldown"] = ctx["currentTick"].(int) + 3 // 3 tick 冷却
		return Success
	})

	// 普通攻击
	normalAttack := NewAction(func(ctx Context) Result {
		ctx["skillUsed"] = false
		return Success
	})

	// 行为树：技能冷却好就用技能，否则普通攻击
	skillTree := NewSelector(
		NewSequence(cooldownReady, useSkillWithCooldown),
		normalAttack,
	)

	// Tick 0：可以使用技能
	ctx["currentTick"] = 0
	ctx["skillCooldown"] = 0
	result := skillTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.True(t, ctx["skillUsed"].(bool))
	assert.Equal(t, 3, ctx["skillCooldown"])

	// Tick 1：冷却中，普通攻击
	ctx["currentTick"] = 1
	ctx["skillUsed"] = false
	result = skillTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.False(t, ctx["skillUsed"].(bool))

	// Tick 2：冷却中，普通攻击
	ctx["currentTick"] = 2
	result = skillTree.Tick(ctx)
	assert.False(t, ctx["skillUsed"].(bool))

	// Tick 3：冷却完成，可以使用技能
	ctx["currentTick"] = 3
	ctx["skillUsed"] = false
	result = skillTree.Tick(ctx)
	assert.True(t, ctx["skillUsed"].(bool))
	assert.Equal(t, 6, ctx["skillCooldown"])
}

// TestGameAI_WanderBehavior 测试游荡行为
// 场景：NPC 在多个巡逻点之间随机移动
func TestGameAI_WanderBehavior(t *testing.T) {
	ctx := make(Context)

	ctx["visitedPoints"] = []string{}

	// 巡逻点动作
	patrolPoint := func(name string) Node {
		return NewAction(func(ctx Context) Result {
			visited := ctx["visitedPoints"].([]string)
			ctx["visitedPoints"] = append(visited, name)
			return Success
		})
	}

	// 随机选择巡逻点
	wanderAI := NewRandomSelector(
		patrolPoint("A"),
		patrolPoint("B"),
		patrolPoint("C"),
	)

	// 执行多次，验证随机性
	visitCounts := make(map[string]int)
	for i := 0; i < 10; i++ {
		ctx["visitedPoints"] = []string{}
		result := wanderAI.Tick(ctx)
		assert.Equal(t, Success, result)

		visited := ctx["visitedPoints"].([]string)
		assert.Len(t, visited, 1)
		visitCounts[visited[0]]++
	}

	// 验证至少访问了2个不同的点（随机性）
	assert.GreaterOrEqual(t, len(visitCounts), 2)
}

// TestGameAI_RetryAttack 测试重试攻击
// 场景：攻击可能被闪避，需要重试
func TestGameAI_RetryAttack(t *testing.T) {
	ctx := make(Context)

	ctx["attackAttempts"] = 0
	ctx["attackSuccess"] = false

	// 可能被闪避的攻击
	unreliableAttack := NewAction(func(ctx Context) Result {
		ctx["attackAttempts"] = ctx["attackAttempts"].(int) + 1
		// 前2次失败，第3次成功
		if ctx["attackAttempts"].(int) >= 3 {
			ctx["attackSuccess"] = true
			return Success
		}
		return Failure
	})

	// 使用 Retry 装饰器：最多重试5次
	attackWithRetry := NewRetry(5, unreliableAttack)

	// 执行攻击
	for {
		result := attackWithRetry.Tick(ctx)
		if result != Running {
			break
		}
	}

	assert.True(t, ctx["attackSuccess"].(bool))
	assert.Equal(t, 3, ctx["attackAttempts"])
}

// TestGameAI_TimeoutEscape 测试超时逃跑
// 场景：尝试逃脱，但如果时间太长就放弃
func TestGameAI_TimeoutEscape(t *testing.T) {
	ctx := make(Context)

	ctx["escapeAttempts"] = 0
	ctx["escaped"] = false

	// 尝试逃脱（会持续尝试）
	tryEscape := NewAction(func(ctx Context) Result {
		ctx["escapeAttempts"] = ctx["escapeAttempts"].(int) + 1
		// 模拟逃脱需要时间
		if ctx["escapeAttempts"].(int) >= 10 {
			ctx["escaped"] = true
			return Success
		}
		return Running
	})

	// 使用 Timeout：最多允许 5 次 tick
	escapeWithTimeout := NewTimeout(5*time.Millisecond, tryEscape)

	// 模拟多次 tick
	result := Running
	for i := 0; i < 100 && result == Running; i++ {
		result = escapeWithTimeout.Tick(ctx)
		time.Sleep(1 * time.Millisecond)
	}

	// 应该因为超时而失败
	assert.Equal(t, Failure, result)
	assert.False(t, ctx["escaped"].(bool))
	assert.Less(t, ctx["escapeAttempts"].(int), 10)
}

// TestGameAI_UseItemLimiter 测试道具使用限制
// 场景：药水只能使用有限次数
func TestGameAI_UseItemLimiter(t *testing.T) {
	ctx := make(Context)

	ctx["health"] = 100
	ctx["healCount"] = 0

	// 治疗动作
	heal := NewAction(func(ctx Context) Result {
		ctx["health"] = ctx["health"].(int) + 20
		ctx["healCount"] = ctx["healCount"].(int) + 1
		return Success
	})

	// 限制只能使用3次治疗药水
	limitedHeal := NewLimiter(3, heal)

	// 使用3次，应该成功
	for i := 0; i < 3; i++ {
		result := limitedHeal.Tick(ctx)
		assert.Equal(t, Success, result)
	}
	assert.Equal(t, 3, ctx["healCount"])
	assert.Equal(t, 160, ctx["health"])

	// 第4次，应该失败（限制已用完）
	result := limitedHeal.Tick(ctx)
	assert.Equal(t, Failure, result)
	assert.Equal(t, 3, ctx["healCount"]) // 没有增加
}

// TestGameAI_DelayedSkill 测试延迟技能
// 场景：蓄力技能需要延迟后才能释放
func TestGameAI_DelayedSkill(t *testing.T) {
	ctx := make(Context)

	ctx["skillCharged"] = false
	ctx["skillReleased"] = false

	// 蓄力完成动作
	chargeComplete := NewAction(func(ctx Context) Result {
		ctx["skillCharged"] = true
		return Success
	})

	// 释放技能动作
	releaseSkill := NewAction(func(ctx Context) Result {
		ctx["skillReleased"] = true
		return Success
	})

	// 延迟3个 tick 后释放
	chargedSkill := NewDelay(3, NewSequence(chargeComplete, releaseSkill))

	// 前3次 tick：蓄力中
	for i := 0; i < 3; i++ {
		result := chargedSkill.Tick(ctx)
		assert.Equal(t, Running, result)
		assert.False(t, ctx["skillCharged"].(bool))
	}

	// 第4次 tick：释放技能
	result := chargedSkill.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.True(t, ctx["skillCharged"].(bool))
	assert.True(t, ctx["skillReleased"].(bool))
}

// TestGameAI_RepeaterForPatrol 测试重复巡逻
// 场景：NPC 重复执行巡逻动作固定次数
func TestGameAI_RepeaterForPatrol(t *testing.T) {
	ctx := make(Context)

	ctx["patrolCount"] = 0

	// 巡逻动作
	patrol := NewAction(func(ctx Context) Result {
		ctx["patrolCount"] = ctx["patrolCount"].(int) + 1
		return Success
	})

	// 重复巡逻5次
	repeater := NewRepeater(5, patrol)

	// 执行直到完成
	for {
		result := repeater.Tick(ctx)
		if result != Running {
			break
		}
	}

	assert.Equal(t, 5, ctx["patrolCount"])
}

// TestGameAI_UntilSuccessPatrol 测试直到成功才停止的巡逻
// 场景：NPC 巡逻直到发现目标
func TestGameAI_UntilSuccessPatrol(t *testing.T) {
	ctx := make(Context)

	ctx["patrolAttempts"] = 0
	ctx["targetFound"] = false

	// 搜索目标动作
	searchTarget := NewAction(func(ctx Context) Result {
		ctx["patrolAttempts"] = ctx["patrolAttempts"].(int) + 1
		// 第4次找到目标
		if ctx["patrolAttempts"].(int) >= 4 {
			ctx["targetFound"] = true
			return Success
		}
		return Failure
	})

	// 直到成功才停止
	untilSuccess := NewUntilSuccess(searchTarget)

	// 执行直到成功
	for {
		result := untilSuccess.Tick(ctx)
		if result == Success {
			break
		}
	}

	assert.True(t, ctx["targetFound"].(bool))
	assert.Equal(t, 4, ctx["patrolAttempts"])
}

// TestGameAI_UntilFailureGuard 测试直到失败才停止的守卫
// 场景：守卫持续警戒直到发现异常
func TestGameAI_UntilFailureGuard(t *testing.T) {
	ctx := make(Context)

	ctx["guardTicks"] = 0
	ctx["alertFound"] = false

	// 警戒检查
	guardCheck := NewAction(func(ctx Context) Result {
		ctx["guardTicks"] = ctx["guardTicks"].(int) + 1
		// 第5次发现异常（返回失败表示需要处理）
		if ctx["guardTicks"].(int) >= 5 {
			ctx["alertFound"] = true
			return Failure
		}
		return Success // 正常，继续警戒
	})

	// 直到失败才停止
	untilFailure := NewUntilFailure(guardCheck)

	// 执行直到失败
	for {
		result := untilFailure.Tick(ctx)
		if result == Success {
			break
		}
	}

	assert.True(t, ctx["alertFound"].(bool))
	assert.Equal(t, 5, ctx["guardTicks"])
}

// TestGameAI_InverterCombatLogic 测试反转战斗逻辑
// 场景：当"没有敌人"时执行休息
func TestGameAI_InverterCombatLogic(t *testing.T) {
	ctx := make(Context)

	ctx["hasEnemy"] = false
	ctx["action"] = ""

	// 检查是否有敌人
	checkEnemy := NewCondition(func(ctx Context) bool {
		return ctx["hasEnemy"].(bool)
	})

	// 休息动作
	rest := NewAction(func(ctx Context) Result {
		ctx["action"] = "resting"
		return Success
	})

	// 战斗动作
	fight := NewAction(func(ctx Context) Result {
		ctx["action"] = "fighting"
		return Success
	})

	// 行为树：有敌人战斗，没有敌人休息
	combatTree := NewSelector(
		NewSequence(checkEnemy, fight),
		NewSequence(NewInverter(checkEnemy), rest),
	)

	// 场景1：没有敌人，休息
	result := combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "resting", ctx["action"])

	// 场景2：有敌人，战斗
	ctx["hasEnemy"] = true
	ctx["action"] = ""
	result = combatTree.Tick(ctx)
	assert.Equal(t, Success, result)
	assert.Equal(t, "fighting", ctx["action"])
}
