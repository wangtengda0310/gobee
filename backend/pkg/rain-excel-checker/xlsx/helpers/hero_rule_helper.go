// Package helpers 提供校验规则的内部辅助工具
// 本包包含列检查、参数解析、表查找等通用辅助函数
package helpers

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
)

// ==================== 数据结构定义 ====================

// ItemConfig 物品配置
type ItemConfig struct {
	ItemId int
	Count  int
}

// HeroRow 武将行数据
type HeroRow struct {
	Id       int
	Name     string
	HeroType int // 武将类型：1=普通武将，其他为特殊武将
	IsOpen   bool
	OpenDate time.Time
	CanMelt  bool
	Skills   []string
	RowIndex int
}

// SeasonPassHero 战令武将信息
type SeasonPassHero struct {
	HeroId       int
	HeroName     string
	SeasonPassId int
	SeasonName   string
	StartTime    time.Time
	EndTime      time.Time
	RowIndex     int
}

// GeneralHero 大将军武将信息
type GeneralHero struct {
	HeroId          int
	HeroName        string
	SeasonId        int
	SeasonName      string
	SeasonStartTime time.Time
	SeasonEndTime   time.Time
	Dan             int
	DanName         string
	RowIndex        int
}

// DropItemInfo 掉落项信息
type DropItemInfo struct {
	Id         int
	Name       string
	DropGroup  int
	Items      []*ItemConfig
	ValidDate  time.Time
	ExpireDate time.Time
	RowIndex   int
}

// SkillMeltInfo 技能熔炼信息
type SkillMeltInfo struct {
	Id        string
	CanMelt   bool
	MeltPower int
	RowIndex  int
}

// ==================== 时间解析工具 ====================

// 通用日期格式列表
// 使用 _1/_2 而非 _01/_02 让 Go time.Parse 兼容补零（"06"）和不补零（"6"）两种形式
var DateFormats = []string{
	"2006-1-2 15:04:05",
	"2006/1/2 15:04:05",
	"2006-1-2",
	"2006/1/2",
}

// ParseDate 解析日期字符串（支持多种格式）
func ParseDate(dateStr string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}
	}

	for _, format := range DateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

// ParseDateWithDefault 使用默认格式解析日期
func ParseDateWithDefault(dateStr, defaultFormat string) time.Time {
	dateStr = strings.TrimSpace(dateStr)
	if dateStr == "" {
		return time.Time{}
	}
	t, err := time.Parse(defaultFormat, dateStr)
	if err != nil {
		return time.Time{}
	}
	return t
}

// TimeEquals 判断两个时间是否精确相等（精确到秒）
func TimeEquals(date1, date2 time.Time) bool {
	if date1.IsZero() || date2.IsZero() {
		return false
	}
	return date1.Format("2006-01-02 15:04:05") == date2.Format("2006-01-02 15:04:05")
}

// TimeIsZero 判断时间是否为零值
func TimeIsZero(t time.Time) bool {
	return t.IsZero()
}

// TimeIsBefore 判断时间是否在指定时间之前
func TimeIsBefore(t, target time.Time) bool {
	if t.IsZero() || target.IsZero() {
		return false
	}
	return t.Before(target)
}

// TimeIsAfter 判断时间是否在指定时间之后
func TimeIsAfter(t, target time.Time) bool {
	if t.IsZero() || target.IsZero() {
		return false
	}
	return t.After(target)
}

// TimeIsInRange 判断时间是否在指定范围内
func TimeIsInRange(t, start, end time.Time) bool {
	if t.IsZero() {
		return false
	}
	afterStart := start.IsZero() || t.After(start) || t.Equal(start)
	beforeEnd := end.IsZero() || t.Before(end) || t.Equal(end)
	return afterStart && beforeEnd
}

// ResolveNow 返回注入的时间，零值回退到 time.Now()。
// 用于单元测试注入固定时间，业务代码通过 CheckParam.Now 传递
func ResolveNow(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now()
	}
	return t
}

// ==================== 物品ID解析工具 ====================

// 物品ID解析正则表达式
var drawSkinItemCfgRegex = regexp.MustCompile(`(\d+);(\d+)`)

// ParseDrawSkinItemCfg 解析物品配置字符串 物品ID;数量
// 格式示例: 1010803;1
func ParseDrawSkinItemCfg(itemCfgStr string) []*ItemConfig {
	items := make([]*ItemConfig, 0)
	if itemCfgStr == "" {
		return items
	}

	matches := drawSkinItemCfgRegex.FindAllStringSubmatch(itemCfgStr, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			itemId, _ := strconv.Atoi(match[1])
			count, _ := strconv.Atoi(match[2])
			items = append(items, &ItemConfig{
				ItemId: itemId,
				Count:  count,
			})
		}
	}
	return items
}

// 物品ID解析正则表达式
var itemCfgRegex = regexp.MustCompile(`\{(\d+);(\d+)\}`)

// ParseItemCfg 解析物品配置字符串 {{物品ID;数量}...}
// 格式示例: {1010803;1}{1000011;10}
func ParseItemCfg(itemCfgStr string) []*ItemConfig {
	items := make([]*ItemConfig, 0)
	if itemCfgStr == "" {
		return items
	}

	matches := itemCfgRegex.FindAllStringSubmatch(itemCfgStr, -1)
	for _, match := range matches {
		if len(match) >= 3 {
			itemId, _ := strconv.Atoi(match[1])
			count, _ := strconv.Atoi(match[2])
			items = append(items, &ItemConfig{
				ItemId: itemId,
				Count:  count,
			})
		}
	}
	return items
}

// ParseCommaSeparatedIds 解析逗号分隔的ID字符串
// 格式示例: 1000001,1000002,1000003
// 用途: 解析 DrawSkin 表的 byproduct 字段
func ParseCommaSeparatedIds(idsStr string) []int {
	ids := make([]int, 0)
	if idsStr == "" {
		return ids
	}

	// 按逗号分割
	parts := strings.Split(idsStr, ",")
	for _, part := range parts {
		// 去除空格并转换为整数
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// ExtractHeroIdFromItemCfg 从物品ID提取武将ID
// 规则：物品ID前两位为10表示武将道具，剩余部分为武将ID
// 例如：1010803 -> 10803
func ExtractHeroIdFromItemCfg(itemId int) int {
	// 物品ID格式: PPXXXXX，PP=10表示武将道具
	prefix := itemId / 100000
	if prefix == 10 {
		return itemId % 100000
	}
	return 0
}

// IsHeroItem 判断物品是否为武将道具
// 武将道具ID规则：武将道具ID = 1000000 + 武将ID
// 有效范围：1001000-1099999（对应 HeroID 1000-99999）
// 排除 1000000-1000999 范围，该范围包含普通道具（如 1000023）
func IsHeroItem(itemId int) bool {
	return itemId >= 1001000 && itemId <= 1099999
}

// MakeHeroItemId 根据武将ID生成武将道具ID
// 规则：武将道具ID = 1000000 + 武将ID
func MakeHeroItemId(heroId int) int {
	return 1000000 + heroId
}

// ==================== 列数据读取工具 ====================

// GetColIndexByName 根据列名获取列索引
func GetColIndexByName(cols [][]string, colName string) int {
	for i, col := range cols {
		if len(col) > excelio.MJS_FIXED_ROWS_NAME && col[excelio.MJS_FIXED_ROWS_NAME] == colName {
			return i
		}
	}
	return -1
}

// GetColValue 获取指定列指定行的值
func GetColValue(cols [][]string, colIndex, rowIndex int) string {
	if colIndex < 0 || colIndex >= len(cols) {
		return ""
	}
	col := cols[colIndex]
	if rowIndex < 0 || rowIndex >= len(col) {
		return ""
	}
	return strings.TrimSpace(col[rowIndex])
}

// GetColValues 获取指定列从起始行开始的所有值
func GetColValues(cols [][]string, colIndex, startRowIdx int) []string {
	if colIndex < 0 || colIndex >= len(cols) {
		return nil
	}
	col := cols[colIndex]
	if startRowIdx >= len(col) {
		return nil
	}
	return col[startRowIdx:]
}

// ==================== 武将查询工具 ====================

// FindHeroById 根据武将ID查找武将行
//
// 执行流程：
//  1. 查找所需列的索引位置（Id、Name、HeroType、IsOpen、OpenDate、CanMelt、Skill）
//  2. 检查Id列是否存在，不存在则返回nil
//  3. 遍历数据行，从起始行开始：
//     a. 读取当前行的Id值并转换为整数
//     b. 跳过空值或解析失败的行
//     c. 找到匹配的武将ID时，提取所有字段信息
//  4. 返回完整的武将行数据，未找到则返回nil
func FindHeroById(heroId int, heroCols [][]string, startRowIdx int) *HeroRow {
	// 步骤1: 查找所需列的索引位置
	idColIdx := GetColIndexByName(heroCols, "Id")
	nameColIdx := GetColIndexByName(heroCols, "Name")
	heroTypeColIdx := GetColIndexByName(heroCols, "HeroType")
	isOpenColIdx := GetColIndexByName(heroCols, "IsOpen")
	openDateColIdx := GetColIndexByName(heroCols, "OpenDate")
	canMeltColIdx := GetColIndexByName(heroCols, "CanMelt")
	skillColIdx := GetColIndexByName(heroCols, "Skill")

	// 步骤2: 检查Id列是否存在
	if idColIdx < 0 {
		return nil
	}

	// 步骤3: 遍历数据行查找匹配的武将ID
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(heroCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取当前行的Id值并转换为整数
		idStr := GetColValue(heroCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		// 步骤3c: 找到匹配的武将ID，提取所有字段信息
		if id == heroId {
			hero := &HeroRow{
				Id:       heroId,
				RowIndex: rowIdx,
			}

			// 提取各列的值（如果列存在）
			if nameColIdx >= 0 {
				hero.Name = GetColValue(heroCols, nameColIdx, rowIdx)
			}
			if heroTypeColIdx >= 0 {
				hero.HeroType, _ = strconv.Atoi(GetColValue(heroCols, heroTypeColIdx, rowIdx))
			}
			if isOpenColIdx >= 0 {
				hero.IsOpen = parseBool(GetColValue(heroCols, isOpenColIdx, rowIdx))
			}
			if openDateColIdx >= 0 {
				hero.OpenDate = ParseDate(GetColValue(heroCols, openDateColIdx, rowIdx))
			}
			if canMeltColIdx >= 0 {
				hero.CanMelt = parseBool(GetColValue(heroCols, canMeltColIdx, rowIdx))
			}
			if skillColIdx >= 0 {
				hero.Skills = parseSkillList(GetColValue(heroCols, skillColIdx, rowIdx))
			}

			// 步骤4: 返回完整的武将行数据
			return hero
		}
	}
	// 未找到匹配的武将ID
	return nil
}

// IsHeroOpened 判断武将是否已开放
// 规则：IsOpen=true 且 OpenDate已过
func IsHeroOpened(hero *HeroRow, now time.Time) bool { // now: 注入的当前时间（零值使用 time.Now()），用于单元测试
	if hero == nil {
		return false
	}
	if !hero.IsOpen {
		return false
	}
	if hero.OpenDate.IsZero() {
		return false
	}
	return hero.OpenDate.Before(ResolveNow(now))
}

// parseBool 解析布尔值
func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

// parseSkillList 解析技能列表
// 格式示例: [Skill1,Skill2] 或 Skill1,Skill2
func parseSkillList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	// 去除方括号
	s = strings.Trim(s, "[]")

	// 按逗号分割
	parts := strings.Split(s, ",")
	skills := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			skills = append(skills, part)
		}
	}
	return skills
}

// DefaultProtectMonths 保护期默认月数（从活动开始时间算起）
const DefaultProtectMonths = 4

// CalcSeasonPassProtectionDeadline 计算战令武将保护期截止时间
// 规则：SeasonPass.StartTime + protectMonths
func CalcSeasonPassProtectionDeadline(startTime time.Time, protectMonths int) time.Time {
	if startTime.IsZero() || protectMonths <= 0 {
		return time.Time{}
	}
	return startTime.AddDate(0, protectMonths, 0)
}

// CalcArenaProtectionDeadline 计算大将军武将保护期截止时间
// 规则：保护期截止时间 = ArenaSeason.SeasonEndTime（赛季结束时间即保护期结束）
// 大将军赛季时长约1个月，赛季结束后武将不再受保护
func CalcArenaProtectionDeadline(seasonEndTime time.Time) time.Time {
	return seasonEndTime
}

// ==================== 战令相关工具 ====================

// FindSeasonPassHeroes 查找所有战令武将
// 规则：按SeasonPassReward表行顺序，找到第一个HighReward包含武将道具的数据
//
// 执行流程：
//  1. 查找SeasonPassReward表中所需列的索引位置
//  2. 初始化去重map和结果切片
//  3. 遍历SeasonPassReward表的数据行：
//     a. 读取HighReward列的值
//     b. 解析其中的物品配置（格式：{物品ID;数量}）
//     c. 检查每个物品是否为武将道具
//     d. 如果是武将道具且未重复，提取武将ID并记录
//     e. 获取对应的战令时间信息
//  4. 返回所有战令武将列表
//
// 参数：
//   - seasonPassRewardCols: SeasonPassReward表的列数据
//   - seasonPassCols: SeasonPass表的列数据（用于获取时间）
//   - startRowIdx: 数据起始行索引
func FindSeasonPassHeroes(seasonPassRewardCols, seasonPassCols [][]string, startRowIdx int) []*SeasonPassHero {
	// 步骤1: 查找列索引
	seasonPassIdColIdx := GetColIndexByName(seasonPassRewardCols, "SeasonPassId")
	highRewardColIdx := GetColIndexByName(seasonPassRewardCols, "HighReward")

	if highRewardColIdx < 0 {
		return make([]*SeasonPassHero, 0)
	}

	// 步骤2: 初始化去重map和结果切片
	heroes := make([]*SeasonPassHero, 0)
	foundHeroIds := make(map[int]bool)

	// 步骤3: 遍历SeasonPassReward表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(seasonPassRewardCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取HighReward列的值
		highReward := GetColValue(seasonPassRewardCols, highRewardColIdx, rowIdx)
		if highReward == "" {
			continue
		}

		// 步骤3b: 解析其中的物品配置
		items := ParseItemCfg(highReward)
		for _, item := range items {
			// 步骤3c: 检查是否是武将道具
			if IsHeroItem(item.ItemId) {
				heroId := ExtractHeroIdFromItemCfg(item.ItemId)

				// 步骤3d: 如果是武将道具且未重复，提取武将ID并记录
				if heroId > 0 && !foundHeroIds[heroId] {
					// 获取战令信息
					seasonPassId := 0
					if seasonPassIdColIdx >= 0 {
						seasonPassId, _ = strconv.Atoi(GetColValue(seasonPassRewardCols, seasonPassIdColIdx, rowIdx))
					}

					hero := &SeasonPassHero{
						HeroId:       heroId,
						SeasonPassId: seasonPassId,
						RowIndex:     rowIdx,
					}

					// 步骤3e: 获取对应的战令时间信息
					if seasonPassCols != nil && seasonPassId > 0 {
						startTime, endTime := GetSeasonPassTime(seasonPassId, seasonPassCols, startRowIdx)
						hero.StartTime = startTime
						hero.EndTime = endTime
					}

					heroes = append(heroes, hero)
					foundHeroIds[heroId] = true
					break // 只取第一个武将
				}
			}
		}
	}

	// 步骤4: 返回所有战令武将列表
	return heroes
}

// GetSeasonPassTime 获取战令时间
//
// 执行流程：
//  1. 查找SeasonPass表中所需列的索引位置
//  2. 检查Id列是否存在，不存在则返回零值
//  3. 遍历SeasonPass表的数据行：
//     a. 读取Id列的值并转换为整数
//     b. 跳过空值或解析失败的行
//     c. 找到匹配的战令ID时，读取StartTime和EndTime
//  4. 返回开始时间和结束时间，未找到则返回零值
func GetSeasonPassTime(seasonPassId int, seasonPassCols [][]string, startRowIdx int) (start, end time.Time) {
	// 步骤1: 查找列索引
	idColIdx := GetColIndexByName(seasonPassCols, "Id")
	startTimeColIdx := GetColIndexByName(seasonPassCols, "StartTime")
	endTimeColIdx := GetColIndexByName(seasonPassCols, "EndTime")

	// 步骤2: 检查Id列是否存在
	if idColIdx < 0 {
		return
	}

	// 步骤3: 遍历查找指定ID
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(seasonPassCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取Id列的值并转换为整数
		idStr := GetColValue(seasonPassCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		// 步骤3c: 找到匹配的战令ID，读取StartTime和EndTime
		if id == seasonPassId {
			if startTimeColIdx >= 0 {
				start = ParseDate(GetColValue(seasonPassCols, startTimeColIdx, rowIdx))
			}
			if endTimeColIdx >= 0 {
				end = ParseDate(GetColValue(seasonPassCols, endTimeColIdx, rowIdx))
			}
			// 步骤4: 返回开始时间和结束时间
			return
		}
	}
	// 未找到则返回零值
	return
}

// ==================== 大将军相关工具 ====================

// FindArenaGeneralHeroes 查找所有大将军武将
// 规则：从ArenaScoreRewards表筛选DanName包含"大将军"的行
//
// 执行流程：
//  1. 查找ArenaScoreRewards表中所需列的索引位置
//  2. 初始化去重map和结果切片
//  3. 遍历ArenaScoreRewards表的数据行：
//     a. 读取DanName列的值，检查是否包含"大将军"
//     b. 如果不包含则跳过该行
//     c. 读取Reward列的值并解析物品配置
//     d. 检查每个物品是否为武将道具
//     e. 如果是武将道具且未重复，提取武将ID并记录
//     f. 读取Season和Dan列的值
//  4. 返回所有大将军武将列表
func FindArenaGeneralHeroes(arenaScoreRewardsCols [][]string, startRowIdx int) []*GeneralHero {
	// 步骤1: 查找列索引
	seasonColIdx := GetColIndexByName(arenaScoreRewardsCols, "Season")
	danColIdx := GetColIndexByName(arenaScoreRewardsCols, "Dan")
	danNameColIdx := GetColIndexByName(arenaScoreRewardsCols, "DanName")
	rewardColIdx := GetColIndexByName(arenaScoreRewardsCols, "Reward")

	if danNameColIdx < 0 || rewardColIdx < 0 {
		return make([]*GeneralHero, 0)
	}

	// 步骤2: 初始化去重map和结果切片
	heroes := make([]*GeneralHero, 0)
	foundHeroIds := make(map[int]bool)

	// 步骤3: 遍历ArenaScoreRewards表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(arenaScoreRewardsCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取DanName列的值，检查是否包含"大将军"
		danName := GetColValue(arenaScoreRewardsCols, danNameColIdx, rowIdx)
		if !strings.Contains(danName, "大将军") {
			// 步骤3b: 不包含则跳过该行
			continue
		}

		// 步骤3c: 读取Reward列的值并解析物品配置
		reward := GetColValue(arenaScoreRewardsCols, rewardColIdx, rowIdx)
		if reward == "" {
			continue
		}

		items := ParseItemCfg(reward)
		for _, item := range items {
			// 步骤3d: 检查每个物品是否为武将道具
			if IsHeroItem(item.ItemId) {
				heroId := ExtractHeroIdFromItemCfg(item.ItemId)

				// 步骤3e: 如果是武将道具且未重复，提取武将ID并记录
				if heroId > 0 && !foundHeroIds[heroId] {
					hero := &GeneralHero{
						HeroId:   heroId,
						DanName:  danName,
						RowIndex: rowIdx,
					}

					// 步骤3f: 读取Season和Dan列的值
					if seasonColIdx >= 0 {
						hero.SeasonId, _ = strconv.Atoi(GetColValue(arenaScoreRewardsCols, seasonColIdx, rowIdx))
					}
					if danColIdx >= 0 {
						hero.Dan, _ = strconv.Atoi(GetColValue(arenaScoreRewardsCols, danColIdx, rowIdx))
					}

					heroes = append(heroes, hero)
					foundHeroIds[heroId] = true
					break
				}
			}
		}
	}

	// 步骤4: 返回所有大将军武将列表
	return heroes
}

// FindGeneralDan 查找大将军段位的Dan值
func FindGeneralDan(arenaScoreRewardsCols [][]string, startRowIdx int) int {
	danColIdx := GetColIndexByName(arenaScoreRewardsCols, "Dan")
	danNameColIdx := GetColIndexByName(arenaScoreRewardsCols, "DanName")

	if danNameColIdx < 0 || danColIdx < 0 {
		return -1
	}

	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(arenaScoreRewardsCols, startRowIdx); rowIdx++ {
		danName := GetColValue(arenaScoreRewardsCols, danNameColIdx, rowIdx)
		if strings.Contains(danName, "大将军") {
			dan, _ := strconv.Atoi(GetColValue(arenaScoreRewardsCols, danColIdx, rowIdx))
			return dan
		}
	}
	return -1
}

// GetArenaSeasonTime 获取竞技场赛季时间
//
// 执行流程：
//  1. 查找ArenaSeason表中所需列的索引位置
//  2. 检查Id列是否存在，不存在则返回零值
//  3. 遍历ArenaSeason表的数据行：
//     a. 读取Id列的值并转换为整数
//     b. 跳过空值或解析失败的行
//     c. 找到匹配的赛季ID时，读取SeasonStartTime和SeasonEndTime
//  4. 返回开始时间和结束时间，未找到则返回零值
func GetArenaSeasonTime(seasonId int, arenaSeasonCols [][]string, startRowIdx int) (start, end time.Time) {
	// 步骤1: 查找列索引
	idColIdx := GetColIndexByName(arenaSeasonCols, "Id")
	startTimeColIdx := GetColIndexByName(arenaSeasonCols, "SeasonStartTime")
	endTimeColIdx := GetColIndexByName(arenaSeasonCols, "SeasonEndTime")

	// 步骤2: 检查Id列是否存在
	if idColIdx < 0 {
		return
	}

	// 步骤3: 遍历查找指定ID
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(arenaSeasonCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取Id列的值并转换为整数
		idStr := GetColValue(arenaSeasonCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		// 步骤3c: 找到匹配的赛季ID，读取SeasonStartTime和SeasonEndTime
		if id == seasonId {
			if startTimeColIdx >= 0 {
				start = ParseDate(GetColValue(arenaSeasonCols, startTimeColIdx, rowIdx))
			}
			if endTimeColIdx >= 0 {
				end = ParseDate(GetColValue(arenaSeasonCols, endTimeColIdx, rowIdx))
			}
			// 步骤4: 返回开始时间和结束时间
			return
		}
	}
	// 未找到则返回零值
	return
}

// GetNextArenaSeasonStartTime 获取下一个即将开始的 ArenaSeason 的开始时间
// 遍历 ArenaSeason 表，找到 SeasonStartTime > now 且最早的记录
//
// 参数：
//   - arenaSeasonCols: ArenaSeason 表的列数据
//   - startRowIdx: 数据起始行索引
//   - now: 当前时间
//
// 返回：下一个赛季的开始时间（未找到则返回零值）
func GetNextArenaSeasonStartTime(arenaSeasonCols [][]string, startRowIdx int, now time.Time) time.Time {
	idColIdx := GetColIndexByName(arenaSeasonCols, "Id")
	startTimeColIdx := GetColIndexByName(arenaSeasonCols, "SeasonStartTime")

	if idColIdx < 0 || startTimeColIdx < 0 {
		return time.Time{}
	}

	var nextStart time.Time
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(arenaSeasonCols, startRowIdx); rowIdx++ {
		startStr := GetColValue(arenaSeasonCols, startTimeColIdx, rowIdx)
		if startStr == "" {
			continue
		}
		startTime := ParseDate(startStr)
		if startTime.IsZero() {
			continue
		}
		// 找到未来开始且最早的赛季
		if startTime.After(now) && (nextStart.IsZero() || startTime.Before(nextStart)) {
			nextStart = startTime
		}
	}
	return nextStart
}

// ==================== 掉落相关工具 ====================

// HeroDropPoolStatus 武将掉落库状态
type HeroDropPoolStatus int

const (
	// HeroNotInDropPool 武将未加入掉落库（DropItem表中无配置）
	HeroNotInDropPool HeroDropPoolStatus = iota
	// HeroInDropPool 武将在掉落库中且时间有效
	HeroInDropPool
	// HeroDropConfigNotEffective 武将已配置掉落但未生效（ValidDate在未来）
	HeroDropConfigNotEffective
	// HeroDropConfigExpired 武将已配置掉落但已过期（ExpireDate已过）
	HeroDropConfigExpired
)

// GetHeroDropPoolStatus 获取武将在掉落库中的状态
// 返回状态类型和详细信息（如生效时间、过期时间等）
//
// 执行流程：
//  1. 根据武将ID计算武将道具ID（武将道具ID = 1000000 + 武将ID）
//  2. 查找DropItem表中所需列的索引位置
//  3. 遍历DropItem表的数据行：
//     a. 读取Item列的值并解析物品配置
//     b. 检查是否包含目标武将道具
//     c. 如果包含，读取ValidDate和ExpireDate
//     d. 判断时间状态：
//     - ValidDate ≤ now ≤ ExpireDate：在掉落库中
//     - ValidDate > now：已配置但未生效
//     - ExpireDate < now：已配置但已过期
//     - 其他情况：未加入掉落库
//  4. 返回状态类型和时间信息
func GetHeroDropPoolStatus(heroId int, dropItemCols [][]string, startRowIdx int, now time.Time) (HeroDropPoolStatus, string, string) { // now: 注入的当前时间（零值使用 time.Now()），用于单元测试
	// 步骤1: 根据武将ID计算武将道具ID
	heroItemId := MakeHeroItemId(heroId)
	now = ResolveNow(now)

	// 步骤2: 查找列索引
	itemColIdx := GetColIndexByName(dropItemCols, "Item")
	validDateColIdx := GetColIndexByName(dropItemCols, "ValidDate")
	expireDateColIdx := GetColIndexByName(dropItemCols, "ExpireDate")

	if itemColIdx < 0 {
		return HeroNotInDropPool, "", ""
	}

	// 步骤3: 遍历DropItem表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(dropItemCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取Item列的值并解析物品配置
		itemCfg := GetColValue(dropItemCols, itemColIdx, rowIdx)
		if itemCfg == "" {
			continue
		}

		items := ParseItemCfg(itemCfg)
		for _, item := range items {
			// 步骤3b: 检查是否包含目标武将道具
			if item.ItemId == heroItemId {
				// 步骤3c: 读取ValidDate和ExpireDate
				validDate := ParseDate(GetColValue(dropItemCols, validDateColIdx, rowIdx))
				expireDate := ParseDate(GetColValue(dropItemCols, expireDateColIdx, rowIdx))

				// 步骤3d: 判断时间状态
				// 状态1: 在掉落库中且时间有效
				if TimeIsInRange(now, validDate, expireDate) {
					return HeroInDropPool, "", ""
				}

				// 状态2: 已配置掉落但未生效（ValidDate在未来）
				if !validDate.IsZero() && validDate.After(now) {
					return HeroDropConfigNotEffective, FormatDateTime(validDate), FormatDateTime(expireDate)
				}

				// 状态3: 已配置掉落但已过期（ExpireDate已过）
				if !expireDate.IsZero() && expireDate.Before(now) {
					return HeroDropConfigExpired, FormatDateTime(validDate), FormatDateTime(expireDate)
				}

				// 状态4: 时间都为空或其他情况
				return HeroNotInDropPool, "", ""
			}
		}
	}

	// 步骤4: 未找到武将道具配置
	return HeroNotInDropPool, "", ""
}

// IsHeroInDropPool 判断武将是否在掉落库中
// 规则：检查DropItem.Item字段，并验证ValidDate/ExpireDate
//
// 执行流程：
//  1. 根据武将ID计算武将道具ID（武将道具ID = 1000000 + 武将ID）
//  2. 获取当前时间
//  3. 查找DropItem表中所需列的索引位置
//  4. 检查Item列是否存在，不存在则返回false
//  5. 遍历DropItem表的数据行：
//     a. 读取Item列的值并解析物品配置
//     b. 检查是否包含目标武将道具
//     c. 如果包含，读取ValidDate和ExpireDate
//     d. 判断当前时间是否在有效范围内（ValidDate ≤ now ≤ ExpireDate）
//     e. 在范围内则返回true
//  6. 未找到或不在范围内则返回false
func IsHeroInDropPool(heroId int, dropItemCols [][]string, startRowIdx int, now time.Time) bool { // now: 注入的当前时间（零值使用 time.Now()），用于单元测试
	// 步骤1: 根据武将ID计算武将道具ID
	heroItemId := MakeHeroItemId(heroId)

	// 步骤2: 获取当前时间
	now = ResolveNow(now)

	// 步骤3: 查找列索引
	itemColIdx := GetColIndexByName(dropItemCols, "Item")
	validDateColIdx := GetColIndexByName(dropItemCols, "ValidDate")
	expireDateColIdx := GetColIndexByName(dropItemCols, "ExpireDate")

	// 步骤4: 检查Item列是否存在
	if itemColIdx < 0 {
		return false
	}

	// 步骤5: 遍历DropItem表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(dropItemCols, startRowIdx); rowIdx++ {
		// 步骤5a: 读取Item列的值并解析物品配置
		itemCfg := GetColValue(dropItemCols, itemColIdx, rowIdx)
		if itemCfg == "" {
			continue
		}

		items := ParseItemCfg(itemCfg)
		for _, item := range items {
			// 步骤5b: 检查是否包含目标武将道具
			if item.ItemId == heroItemId {
				// 步骤5c: 读取ValidDate和ExpireDate
				validDate := ParseDate(GetColValue(dropItemCols, validDateColIdx, rowIdx))
				expireDate := ParseDate(GetColValue(dropItemCols, expireDateColIdx, rowIdx))

				// 步骤5d: 判断当前时间是否在有效范围内
				// ValidDate ≤ now ≤ ExpireDate
				if TimeIsInRange(now, validDate, expireDate) {
					// 步骤5e: 在范围内则返回true
					return true
				}
			}
		}
	}

	// 步骤6: 未找到或不在范围内则返回false
	return false
}

// GetDropItemInfos 获取所有掉落项信息
//
// 执行流程：
//  1. 查找DropItem表中所需列的索引位置
//  2. 检查Item列是否存在，不存在则返回空切片
//  3. 遍历DropItem表的数据行：
//     a. 读取Item列的值，跳过空值
//     b. 创建掉落项信息结构体
//     c. 解析Item列的物品配置
//     d. 读取ValidDate和ExpireDate列的值
//     e. 读取Id、Name、DropGroup列的值
//     f. 将掉落项信息加入结果切片
//  4. 返回所有掉落项信息列表
func GetDropItemInfos(dropItemCols [][]string, startRowIdx int) []*DropItemInfo {
	// 步骤1: 查找列索引
	idColIdx := GetColIndexByName(dropItemCols, "Id")
	nameColIdx := GetColIndexByName(dropItemCols, "Name")
	dropGroupColIdx := GetColIndexByName(dropItemCols, "DropGroup")
	itemColIdx := GetColIndexByName(dropItemCols, "Item")
	validDateColIdx := GetColIndexByName(dropItemCols, "ValidDate")
	expireDateColIdx := GetColIndexByName(dropItemCols, "ExpireDate")

	// 步骤2: 检查Item列是否存在
	if itemColIdx < 0 {
		return make([]*DropItemInfo, 0)
	}

	// 初始化结果切片
	items := make([]*DropItemInfo, 0)

	// 步骤3: 遍历DropItem表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(dropItemCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取Item列的值，跳过空值
		itemCfg := GetColValue(dropItemCols, itemColIdx, rowIdx)
		if itemCfg == "" {
			continue
		}

		// 步骤3b: 创建掉落项信息结构体
		info := &DropItemInfo{
			RowIndex: rowIdx,
			// 步骤3c: 解析Item列的物品配置
			Items: ParseItemCfg(itemCfg),
			// 步骤3d: 读取ValidDate和ExpireDate列的值
			ValidDate:  ParseDate(GetColValue(dropItemCols, validDateColIdx, rowIdx)),
			ExpireDate: ParseDate(GetColValue(dropItemCols, expireDateColIdx, rowIdx)),
		}

		// 步骤3e: 读取Id、Name、DropGroup列的值
		if idColIdx >= 0 {
			info.Id, _ = strconv.Atoi(GetColValue(dropItemCols, idColIdx, rowIdx))
		}
		if nameColIdx >= 0 {
			info.Name = GetColValue(dropItemCols, nameColIdx, rowIdx)
		}
		if dropGroupColIdx >= 0 {
			info.DropGroup, _ = strconv.Atoi(GetColValue(dropItemCols, dropGroupColIdx, rowIdx))
		}

		// 步骤3f: 将掉落项信息加入结果切片
		items = append(items, info)
	}

	// 步骤4: 返回所有掉落项信息列表
	return items
}

// ==================== 合成相关工具 ====================

// IsHeroSynthesisEnabled 判断武将是否可合成，并返回 Item 表中对应道具的行索引
// 规则：检查Item表的IsSynthetic字段
// 返回值：(isSynthetic, itemRowIdx)，找不到时返回 (false, 0)
func IsHeroSynthesisEnabled(heroId int, itemCols [][]string, startRowIdx int) (bool, int) {
	heroItemId := MakeHeroItemId(heroId)

	// 查找列索引
	idColIdx := GetColIndexByName(itemCols, "Id")
	isSyntheticColIdx := GetColIndexByName(itemCols, "IsSynthetic")

	if idColIdx < 0 {
		return false, 0
	}

	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(itemCols, startRowIdx); rowIdx++ {
		idStr := GetColValue(itemCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}
		if id == heroItemId {
			if isSyntheticColIdx >= 0 {
				return parseBool(GetColValue(itemCols, isSyntheticColIdx, rowIdx)), rowIdx
			}
			return false, rowIdx
		}
	}
	return false, 0
}

// ==================== 熔炼相关工具 ====================

// IsHeroMeltEnabled 判断武将是否可熔炼
// 规则：检查Hero表的CanMelt字段
func IsHeroMeltEnabled(heroId int, heroCols [][]string, startRowIdx int) bool {
	hero := FindHeroById(heroId, heroCols, startRowIdx)
	if hero == nil {
		return false
	}
	return hero.CanMelt
}

// BuildSkillMeltMap 构建技能熔炼映射表
//
// 执行流程：
//  1. 查找SkillMelt表中所需列的索引位置
//  2. 检查Id列是否存在，不存在则返回空map
//  3. 遍历SkillMelt表的数据行：
//     a. 读取技能ID，跳过空值
//     b. 创建技能熔炼信息结构体
//     c. 读取CanMelt和MeltPower列的值
//     d. 将技能信息存入map，以技能ID为key
//  4. 返回完整的技能熔炼映射表
func BuildSkillMeltMap(skillMeltCols [][]string, startRowIdx int) map[string]*SkillMeltInfo {
	// 步骤1: 查找列索引
	idColIdx := GetColIndexByName(skillMeltCols, "Id")
	canMeltColIdx := GetColIndexByName(skillMeltCols, "CanMelt")
	meltPowerColIdx := GetColIndexByName(skillMeltCols, "MeltPower")

	// 步骤2: 检查Id列是否存在
	if idColIdx < 0 {
		return make(map[string]*SkillMeltInfo)
	}

	// 初始化结果map
	meltMap := make(map[string]*SkillMeltInfo)

	// 步骤3: 遍历SkillMelt表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(skillMeltCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取技能ID，跳过空值
		id := GetColValue(skillMeltCols, idColIdx, rowIdx)
		if id == "" {
			continue
		}

		// 步骤3b: 创建技能熔炼信息结构体
		info := &SkillMeltInfo{
			Id:       id,
			RowIndex: rowIdx,
		}

		// 步骤3c: 读取CanMelt和MeltPower列的值
		if canMeltColIdx >= 0 {
			info.CanMelt = parseBool(GetColValue(skillMeltCols, canMeltColIdx, rowIdx))
		}
		if meltPowerColIdx >= 0 {
			info.MeltPower, _ = strconv.Atoi(GetColValue(skillMeltCols, meltPowerColIdx, rowIdx))
		}

		// 步骤3d: 将技能信息存入map
		meltMap[id] = info
	}

	// 步骤4: 返回完整的技能熔炼映射表
	return meltMap
}

// IsSkillMeltConfigured 判断技能是否配置了熔炼
// 规则：检查SkillMelt表是否有该技能的记录且CanMelt=true
func IsSkillMeltConfigured(skillId string, skillMeltMap map[string]*SkillMeltInfo) bool {
	info, exists := skillMeltMap[skillId]
	if !exists {
		return false
	}
	return info.CanMelt
}

// ==================== 通知格式化工具 ====================

// FormatChangeMessage 格式化变更消息
// 格式: "xx行，id为xx，字段y从a改成了b"
func FormatChangeMessage(rowName, rowId, colName, oldValue, newValue string) string {
	return fmt.Sprintf("%s行，id为%s，字段%s从%s改成了%s", rowName, rowId, colName, oldValue, newValue)
}

// FormatAddRowMessage 格式化新增行消息
func FormatAddRowMessage(rowId, rowName string) string {
	return fmt.Sprintf("ID=%s, 名称=%s", rowId, rowName)
}

// FormatDate 格式化日期为标准格式
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// FormatDateTime 格式化日期时间为标准格式
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// IsHeroInSeasonPassReward 判断武将是否在SeasonPassReward表中
// 用于checkOpenedHeroes跳过战令武将，让checkSeasonPassHeroes单独处理
//
// 执行流程：
//  1. 检查SeasonPassReward表数据是否为空
//  2. 查找HighReward列的索引位置
//  3. 遍历SeasonPassReward表的数据行：
//     a. 读取HighReward列的值
//     b. 根据武将ID构造武将道具ID（武将道具ID = 1000000 + 武将ID）
//     c. 构造目标字符串格式：{武将道具ID;
//     d. 检查HighReward值是否包含目标字符串
//     e. 包含则返回true（该武将是战令武将）
//  4. 未找到则返回false
//
// 参数：
//   - heroId: 武将ID
//   - seasonPassRewardCols: SeasonPassReward表的列数据
//   - startRowIdx: 数据起始行索引
//
// 返回：true=战令武将，false=非战令武将
func IsHeroInSeasonPassReward(heroId int, seasonPassRewardCols [][]string, startRowIdx int) bool {
	// 步骤1: 检查SeasonPassReward表数据是否为空
	if seasonPassRewardCols == nil {
		return false
	}

	// 步骤2: 查找HighReward列的索引位置
	highRewardColIdx := GetColIndexByName(seasonPassRewardCols, "HighReward")
	if highRewardColIdx < 0 {
		return false
	}

	// 步骤3: 遍历SeasonPassReward表的数据行
	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(seasonPassRewardCols, startRowIdx); rowIdx++ {
		// 步骤3a: 读取HighReward列的值
		highReward := GetColValue(seasonPassRewardCols, highRewardColIdx, rowIdx)
		if highReward == "" {
			continue
		}

		// 步骤3b: 根据武将ID构造武将道具ID
		heroItemId := 1000000 + heroId

		// 步骤3c: 构造目标字符串格式
		target := fmt.Sprintf("{%d;", heroItemId)

		// 步骤3d: 检查HighReward值是否包含目标字符串
		if contains(highReward, target) {
			// 步骤3e: 包含则返回true
			return true
		}
	}

	// 步骤4: 未找到则返回false
	return false
}

// BuildHeroItemIdMap 构建所有武将道具的 ItemId → HeroId 映射
// 使用语义映射: Item.Type == "Hero" → Item.ItemParam = 武将ID
// ItemParam 对 Hero 类型为纯数字字符串（如 "10001"），直接 Atoi 转换
//
// 执行流程：
//  1. 查找 Item 表中的 Id、Type、ItemParam 列索引
//  2. 检查必要列是否存在
//  3. 遍历 Item 表数据行：
//     a. 读取 Type 列值
//     b. 当 Type 包含 "Hero" 时，读取 Id 和 ItemParam
//     c. Atoi 转换 Id 和 ItemParam
//     d. 记录 Id → ItemParam 映射
//  4. 返回映射 map
func BuildHeroItemIdMap(itemCols [][]string, startRowIdx int) map[int]int {
	// 步骤1: 查找列索引
	idColIdx := GetColIndexByName(itemCols, "Id")
	typeColIdx := GetColIndexByName(itemCols, "Type")
	paramColIdx := GetColIndexByName(itemCols, "ItemParam")

	if idColIdx < 0 || typeColIdx < 0 {
		return make(map[int]int)
	}

	// 步骤2: 构建映射
	heroItemMap := make(map[int]int)

	for rowIdx := startRowIdx; rowIdx < GetDataEndIndex(itemCols, startRowIdx); rowIdx++ {
		typeVal := GetColValue(itemCols, typeColIdx, rowIdx)
		if !strings.Contains(typeVal, "Hero") {
			continue
		}

		idStr := GetColValue(itemCols, idColIdx, rowIdx)
		if idStr == "" {
			continue
		}
		itemId, err := strconv.Atoi(idStr)
		if err != nil {
			continue
		}

		if paramColIdx >= 0 {
			paramStr := GetColValue(itemCols, paramColIdx, rowIdx)
			if paramStr == "" {
				continue
			}
			// ItemParam 对 Hero 类型是纯数字（如 "10001"）
			heroId, err := strconv.Atoi(paramStr)
			if err != nil {
				continue
			}
			heroItemMap[itemId] = heroId
		}
	}

	return heroItemMap
}

// contains 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

// findSubstring 辅助函数：在s中查找substr
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
