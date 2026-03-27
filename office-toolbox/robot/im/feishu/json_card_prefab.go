package feishu

type JsonCardPrefab struct{}

// NormalBlue 蓝色信息卡片
func (j *JsonCardPrefab) NormalBlue(title, subTitle string, element ...Element) *JsonCard {
	return &JsonCard{
		Schema: "2.0",
		Config: &Config{
			EnableForward: true,
			UpdateMulti:   true,
		},
		Body: &Body{
			Direction: "vertical",
			Padding:   "12px 12px 12px 12px",
			Elements:  element,
		},
		Header: &Header{
			Title: &TextElement{
				Tag:     "plain_text",
				Content: title,
			},
			Subtitle: &TextElement{
				Tag:     "plain_text",
				Content: subTitle,
			},
			Template: "blue",
			Padding:  "12px 12px 12px 12px",
		},
	}
}

// SuccessGreen 绿色成功卡片
func (j *JsonCardPrefab) SuccessGreen(title, subTitle string, element ...Element) *JsonCard {
	return &JsonCard{
		Schema: "2.0",
		Config: &Config{
			EnableForward: true,
			UpdateMulti:   true,
		},
		Body: &Body{
			Direction: "vertical",
			Padding:   "12px 12px 12px 12px",
			Elements:  element,
		},
		Header: &Header{
			Title: &TextElement{
				Tag:     "plain_text",
				Content: title,
			},
			Subtitle: &TextElement{
				Tag:     "plain_text",
				Content: subTitle,
			},
			Template: "green",
			Padding:  "12px 12px 12px 12px",
		},
	}
}

// WarningRed 红色警告卡片
func (j *JsonCardPrefab) WarningRed(title, subTitle string, element ...Element) *JsonCard {
	return &JsonCard{
		Schema: "2.0",
		Config: &Config{
			EnableForward: true,
			UpdateMulti:   true,
		},
		Body: &Body{
			Direction: "vertical",
			Padding:   "12px 12px 12px 12px",
			Elements:  element,
		},
		Header: &Header{
			Title: &TextElement{
				Tag:     "plain_text",
				Content: title,
			},
			Subtitle: &TextElement{
				Tag:     "plain_text",
				Content: subTitle,
			},
			Template: "red",
			Padding:  "12px 12px 12px 12px",
		},
	}
}
