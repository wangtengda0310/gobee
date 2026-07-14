package shop

import (
	"strconv"
	"strings"

	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/excelio"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-excel-checker/xlsx/helpers"
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/utils"
	"github.com/xuri/excelize/v2"
)

type ShopDiff struct {
	ShopType        string // 商店类型(EShopType枚举)
	Name            string // 商店名称
	ShopName        string // 商店显示名称
	UseCurrency     []int  // 商店使用的货币
	OpenTime        string // 开启时间
	CloseTime       string // 关闭时间
	IsLimitedShop   bool   // 是否为限时商店
	IsDynamicsShop  bool   // 是否为动态商店
	InMainShopOrder int    // 主界面中商店顺序
}

func (s ShopDiff) GetType() string {
	return "ShopDiff"
}

func (s ShopDiff) GetDisplayName() string {
	return s.Name
}

func GetShopDiffMap(sheetMap map[string]*excelize.File) (diffInfo *[]ShopDiff, err error) {
	var shopCols [][]string // 商店表
	if sheet, exist := sheetMap["商店表|Shop"]; exist {
		var err error
		shopCols, err = sheet.GetCols("商店表|Shop")
		if err != nil {
			return nil, err
		}
	}

	// 名将杀专属配置
	startRow := excelio.MJS_FIXED_ROWS_NUM

	// 构建商店Map
	shopsDiff := make([]ShopDiff, 0, 50)

	// 第一列是EShopType枚举字符串，判断规则：不为空且不以#开头
	for i, shopTypeStr := range shopCols[ShopType][startRow:helpers.AutoDetectEndIndex(shopCols, ShopType, startRow, 3)] {
		// 跳过空行和注释行（以#开头）
		if shopTypeStr == "" || strings.HasPrefix(shopTypeStr, "#") {
			continue
		}

		shopDiff := ShopDiff{}

		// 获取各列数据
		name := utils.GetCellValue(shopCols, Name, startRow+i)
		shopName := utils.GetCellValue(shopCols, ShopName, startRow+i)
		useCurrency := utils.GetCellValue(shopCols, UseCurrency, startRow+i)
		openTime := utils.GetCellValue(shopCols, OpenTime, startRow+i)
		closeTime := utils.GetCellValue(shopCols, CloseTime, startRow+i)
		isLimitedShop := utils.GetCellValue(shopCols, IsLimitedShop, startRow+i)
		isDynamicsShop := utils.GetCellValue(shopCols, IsDynamicsShop, startRow+i)
		inMainShopOrder := utils.GetCellValue(shopCols, InMainShopOrder, startRow+i)

		// 赋值
		shopDiff.ShopType = shopTypeStr
		shopDiff.Name = name
		shopDiff.ShopName = shopName

		// 解析UseCurrency（int数组）
		shopDiff.UseCurrency = make([]int, 0, 5)
		for _, s := range strings.Split(useCurrency, ",") {
			if s == "" {
				continue
			}
			if n, err := strconv.Atoi(s); err == nil {
				shopDiff.UseCurrency = append(shopDiff.UseCurrency, n)
			}
		}

		shopDiff.OpenTime = openTime
		shopDiff.CloseTime = closeTime

		if b, err := strconv.ParseBool(isLimitedShop); err == nil {
			shopDiff.IsLimitedShop = b
		} else {
			shopDiff.IsLimitedShop = false
		}

		if b, err := strconv.ParseBool(isDynamicsShop); err == nil {
			shopDiff.IsDynamicsShop = b
		} else {
			shopDiff.IsDynamicsShop = false
		}

		if n, err := strconv.Atoi(inMainShopOrder); err == nil {
			shopDiff.InMainShopOrder = n
		} else {
			shopDiff.InMainShopOrder = -1
		}

		shopsDiff = append(shopsDiff, shopDiff)
	}

	return &shopsDiff, nil
}
