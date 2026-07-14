package shop_goods

import (
	"regexp"
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type ShopGoodsDiff struct {
	Id                     int
	Name                   string
	Desc                   string
	Item                   []ItemCfg
	ExtraItem              []ItemCfg
	CostId                 int
	ShopType               string
	RechargeAndriod        int
	RechargeIOS            int
	RechargeMulti          int
	RechargeMultiGroup     int
	Price                  int
	OldPrice               int
	Discount               int
	LimitType              int
	LimitCount             int
	MaxBuyCount            int
	Pos                    int
	IsNew                  bool
	Countdown              int
	IsHomePage             bool
	Icon                   string
	OnShelfTime            string
	OffShelfTime           string
	RechargeMultiBeginTime string
	RechargeMultiEndTime   string
	PreSelOutShopId        int
	HideWhenSelOut         bool
	LimitZeroWhenHasItem   bool
	IconDisplayType        int
	ShowItemCount          bool
	ShowQuality            bool
	IconInBuyWindow        string
	RechargeMultiGoodsDes  string
	RewardID               int
}

type ItemCfg struct {
	ItemId int
	Count  int
}

func (s ShopGoodsDiff) GetType() string {
	return "ShopGoodsDiff"
}

func (s ShopGoodsDiff) GetDisplayName() string {
	return s.Name
}

func GetShopGoodsDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]ShopGoodsDiff, err error) {
	var shopGoodsCols [][]string // 商品表
	if sheet, exist := sheetMap["商品表|ShopGood"]; exist {
		var err error
		shopGoodsCols, err = sheet.GetCols("商品表|ShopGood")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建商品Map
	shopGoodsDiff := make([]ShopGoodsDiff, 0, 300)

	reg := regexp.MustCompile(`\{(\d+);\d+}`)

	for i, idStr := range shopGoodsCols[Id][startRow:helpers.AutoDetectEndIndex(shopGoodsCols, Id, startRow, 3)] {
		// 表格第一列是int类型的Id
		if id, err := strconv.Atoi(idStr); err != nil {
			continue
		} else {
			goodsDiff := ShopGoodsDiff{}

			name := utils.GetCellValue(shopGoodsCols, Name, startRow+i)
			desc := utils.GetCellValue(shopGoodsCols, Desc, startRow+i)
			item := utils.GetCellValue(shopGoodsCols, Item, startRow+i)
			extraItem := utils.GetCellValue(shopGoodsCols, ExtraItem, startRow+i)
			costId := utils.GetCellValue(shopGoodsCols, CostId, startRow+i)
			shopType := utils.GetCellValue(shopGoodsCols, ShopType, startRow+i)
			rechargeAndriod := utils.GetCellValue(shopGoodsCols, RechargeAndriod, startRow+i)
			rechargeIOS := utils.GetCellValue(shopGoodsCols, RechargeIOS, startRow+i)
			rechargeMulti := utils.GetCellValue(shopGoodsCols, RechargeMulti, startRow+i)
			rechargeMultiGroup := utils.GetCellValue(shopGoodsCols, RechargeMultiGroup, startRow+i)
			price := utils.GetCellValue(shopGoodsCols, Price, startRow+i)
			oldPrice := utils.GetCellValue(shopGoodsCols, OldPrice, startRow+i)
			discount := utils.GetCellValue(shopGoodsCols, Discount, startRow+i)
			limitType := utils.GetCellValue(shopGoodsCols, LimitType, startRow+i)
			limitCount := utils.GetCellValue(shopGoodsCols, LimitCount, startRow+i)
			maxBuyCount := utils.GetCellValue(shopGoodsCols, MaxBuyCount, startRow+i)
			pos := utils.GetCellValue(shopGoodsCols, Pos, startRow+i)
			isNew := utils.GetCellValue(shopGoodsCols, IsNew, startRow+i)
			countdown := utils.GetCellValue(shopGoodsCols, Countdown, startRow+i)
			isHomePage := utils.GetCellValue(shopGoodsCols, IsHomePage, startRow+i)
			icon := utils.GetCellValue(shopGoodsCols, Icon, startRow+i)
			onShelfTime := utils.GetCellValue(shopGoodsCols, OnShelfTime, startRow+i)
			offShelfTime := utils.GetCellValue(shopGoodsCols, OffShelfTime, startRow+i)
			rechargeMultiBeginTime := utils.GetCellValue(shopGoodsCols, RechargeMultiBeginTime, startRow+i)
			rechargeMultiEndTime := utils.GetCellValue(shopGoodsCols, RechargeMultiEndTime, startRow+i)
			preSelOutShopId := utils.GetCellValue(shopGoodsCols, PreSelOutShopId, startRow+i)
			hideWhenSelOut := utils.GetCellValue(shopGoodsCols, HideWhenSelOut, startRow+i)
			limitZeroWhenHasItem := utils.GetCellValue(shopGoodsCols, LimitZeroWhenHasItem, startRow+i)
			iconDisplayType := utils.GetCellValue(shopGoodsCols, IconDisplayType, startRow+i)
			showItemCount := utils.GetCellValue(shopGoodsCols, ShowItemCount, startRow+i)
			showQuality := utils.GetCellValue(shopGoodsCols, ShowQuality, startRow+i)
			iconInBuyWindow := utils.GetCellValue(shopGoodsCols, IconInBuyWindow, startRow+i)
			rechargeMultiGoodsDes := utils.GetCellValue(shopGoodsCols, RechargeMultiGoodsDes, startRow+i)
			rewardID := utils.GetCellValue(shopGoodsCols, RewardID, startRow+i)

			// 录入商品信息
			goodsDiff.Id = id
			goodsDiff.Name = name
			goodsDiff.Desc = desc

			// 解析Item道具
			goodsDiff.Item = parseItemCfg(item, reg)
			goodsDiff.ExtraItem = parseItemCfg(extraItem, reg)

			if n, err := strconv.Atoi(costId); err == nil {
				goodsDiff.CostId = n
			} else {
				goodsDiff.CostId = -1
			}

			goodsDiff.ShopType = shopType

			if n, err := strconv.Atoi(rechargeAndriod); err == nil {
				goodsDiff.RechargeAndriod = n
			} else {
				goodsDiff.RechargeAndriod = -1
			}

			if n, err := strconv.Atoi(rechargeIOS); err == nil {
				goodsDiff.RechargeIOS = n
			} else {
				goodsDiff.RechargeIOS = -1
			}

			if n, err := strconv.Atoi(rechargeMulti); err == nil {
				goodsDiff.RechargeMulti = n
			} else {
				goodsDiff.RechargeMulti = -1
			}

			if n, err := strconv.Atoi(rechargeMultiGroup); err == nil {
				goodsDiff.RechargeMultiGroup = n
			} else {
				goodsDiff.RechargeMultiGroup = -1
			}

			if n, err := strconv.Atoi(price); err == nil {
				goodsDiff.Price = n
			} else {
				goodsDiff.Price = -1
			}

			if n, err := strconv.Atoi(oldPrice); err == nil {
				goodsDiff.OldPrice = n
			} else {
				goodsDiff.OldPrice = -1
			}

			if n, err := strconv.Atoi(discount); err == nil {
				goodsDiff.Discount = n
			} else {
				goodsDiff.Discount = -1
			}

			if n, err := strconv.Atoi(limitType); err == nil {
				goodsDiff.LimitType = n
			} else {
				goodsDiff.LimitType = -1
			}

			if n, err := strconv.Atoi(limitCount); err == nil {
				goodsDiff.LimitCount = n
			} else {
				goodsDiff.LimitCount = -1
			}

			if n, err := strconv.Atoi(maxBuyCount); err == nil {
				goodsDiff.MaxBuyCount = n
			} else {
				goodsDiff.MaxBuyCount = -1
			}

			if n, err := strconv.Atoi(pos); err == nil {
				goodsDiff.Pos = n
			} else {
				goodsDiff.Pos = -1
			}

			if b, err := strconv.ParseBool(isNew); err == nil {
				goodsDiff.IsNew = b
			} else {
				goodsDiff.IsNew = false
			}

			if n, err := strconv.Atoi(countdown); err == nil {
				goodsDiff.Countdown = n
			} else {
				goodsDiff.Countdown = -1
			}

			if b, err := strconv.ParseBool(isHomePage); err == nil {
				goodsDiff.IsHomePage = b
			} else {
				goodsDiff.IsHomePage = false
			}

			goodsDiff.Icon = icon
			goodsDiff.OnShelfTime = onShelfTime
			goodsDiff.OffShelfTime = offShelfTime
			goodsDiff.RechargeMultiBeginTime = rechargeMultiBeginTime
			goodsDiff.RechargeMultiEndTime = rechargeMultiEndTime

			if n, err := strconv.Atoi(preSelOutShopId); err == nil {
				goodsDiff.PreSelOutShopId = n
			} else {
				goodsDiff.PreSelOutShopId = -1
			}

			if b, err := strconv.ParseBool(hideWhenSelOut); err == nil {
				goodsDiff.HideWhenSelOut = b
			} else {
				goodsDiff.HideWhenSelOut = false
			}

			if b, err := strconv.ParseBool(limitZeroWhenHasItem); err == nil {
				goodsDiff.LimitZeroWhenHasItem = b
			} else {
				goodsDiff.LimitZeroWhenHasItem = false
			}

			if n, err := strconv.Atoi(iconDisplayType); err == nil {
				goodsDiff.IconDisplayType = n
			} else {
				goodsDiff.IconDisplayType = -1
			}

			if b, err := strconv.ParseBool(showItemCount); err == nil {
				goodsDiff.ShowItemCount = b
			} else {
				goodsDiff.ShowItemCount = false
			}

			if b, err := strconv.ParseBool(showQuality); err == nil {
				goodsDiff.ShowQuality = b
			} else {
				goodsDiff.ShowQuality = false
			}

			goodsDiff.IconInBuyWindow = iconInBuyWindow
			goodsDiff.RechargeMultiGoodsDes = rechargeMultiGoodsDes

			if n, err := strconv.Atoi(rewardID); err == nil {
				goodsDiff.RewardID = n
			} else {
				goodsDiff.RewardID = -1
			}

			shopGoodsDiff = append(shopGoodsDiff, goodsDiff)
		}
	}

	return &shopGoodsDiff, nil
}

// parseItemCfg 解析道具配置字符串，格式为 {itemId;count}，多个用逗号分隔
func parseItemCfg(itemStr string, reg *regexp.Regexp) []ItemCfg {
	items := make([]ItemCfg, 0, 5)
	if itemStr == "" {
		return items
	}

	// 使用正则找出所有匹配项
	matches := reg.FindAllStringSubmatch(itemStr, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			// 解析完整格式 {itemId;count}
			fullMatch := match[0] // 完整的 {数字;数字}
			itemIdStr := match[1] // 第一个数字

			// 解析数量
			count := 1
			parts := strings.Split(strings.Trim(fullMatch, "{}"), ";")
			if len(parts) >= 2 {
				if c, err := strconv.Atoi(parts[1]); err == nil {
					count = c
				}
			}

			itemId, _ := strconv.Atoi(itemIdStr)

			items = append(items, ItemCfg{
				ItemId: itemId,
				Count:  count,
			})
		}
	}

	return items
}
