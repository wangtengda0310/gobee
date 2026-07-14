package mjs_excel

import (
	"git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/mjs_excel/hero_skin_item"
	"github.com/xuri/excelize/v2"
)

type ImgRefExcel struct {
	SkinIdNameMap   map[int]string // 皮肤枚举INT=>台词枚举
	SkinIdOpenMap   map[int]bool   // 皮肤枚举INT=>皮肤是否开放
	SkinIdPinYinMap map[int]string // 皮肤枚举INT=>拼音
}

func InitImgRefExcel(sheetMap map[string]*excelize.File) (*ImgRefExcel, error) {
	skinIdNameMap, skinIdOpenMap, skinIdPinYinMap, err := hero_skin_item.GetHeroSkinsItemImgMap(sheetMap)
	if err != nil {
		return nil, err
	}

	excel := &ImgRefExcel{
		skinIdNameMap,
		skinIdOpenMap,
		skinIdPinYinMap,
	}
	return excel, nil
}
