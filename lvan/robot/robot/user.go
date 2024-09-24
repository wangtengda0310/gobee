package robot

import ( "gforge/pb" )

type Account struct {
AccountId uint32
OpenId string
Token string
ZoneId uint32
Addr string
TokenEndTime int64
Permission int32
}

// RegisterReq 游戏账号注册 本身游戏账号使用
type RegisterReq struct{
	UserName string json:"u_name" // 用户名
	UserPWD string json:"u_pwd" // 用户密码
}

// RegisterRep 注册或者登录返回
type RegisterRep struct {
	Ret pb.Errno `json:"ret"` // 错误码
AccountID uint32 `json:"accountid"` // 账号ID
OpenID string `json:"openid"` // OpenID
 Token string `json:"token"` // AccountID Token
 Limit int32 `json:"Limit"` // Limit
  TokenEndTime int64 `json:"Tet"` // TokenEndTime
    Permission int32 `json:"P"` // 权限
	 }