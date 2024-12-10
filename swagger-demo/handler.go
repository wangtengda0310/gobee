package main

type Article struct {
}

// @Summary 获取单个文章
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [get]
func (a Article) Get() {

}

// @Summary 获取多个文章
// @Produce json
// @Param name query string false "文章名称"
// @Param tag_id query int false "标签ID"
// @Param state query int false "状态"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles [get]
func (a Article) List() {
	return
}

// @Summary 创建文章
// @Produce json
// @Param tag_id body string true "标签ID"
// @Param title body string true "文章标题"
// @Param desc body string false "文章简述"
// @Param cover_image_url body string true "封面图片地址"
// @Param content body string true "文章内容"
// @Param created_by body int true "创建者"
// @Param state body int false "状态"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles [post]
func (a Article) Create() {

}

// @Summary 更新文章
// @Produce json
// @Param tag_id body string false "标签ID"
// @Param title body string false "文章标题"
// @Param desc body string false "文章简述"
// @Param cover_image_url body string false "封面图片地址"
// @Param content body string false "文章内容"
// @Param modified_by body string true "修改者"
// @Success 200 {object} Article "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [put]
func (a Article) Update() {
	return
}

// @Summary 删除文章
// @Produce  json
// @Param id path int true "文章ID"
// @Success 200 {string} string "成功"
// @Failure 400 {object} string "请求错误"
// @Failure 500 {object} string "内部错误"
// @Router /api/v1/articles/{id} [delete]
func (a Article) Delete() {
	return
}

type User struct {
	ID   uint64 `json:"id" example:"100" description:"User identity"`
	Name string `json:"name" example:"Parvez"`
}

type UsersResponse struct {
	Data []User `json:"users" example:"[{\"id\":100, \"name\":\"Parvez\"}]"`
}

type Error struct {
	Code string `json:"code"`
	Msg  string `json:"msg" skip:"true"`
	// use skip:"true" if you want to skip the field in the documentation spec
}

type ErrorResponse struct {
	ErrorInfo Error `json:"error"`
}
