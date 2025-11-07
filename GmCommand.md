# 获取命令列表（游戏服务端提供）
||
接口地址：/api/v2/kill/role.do
接口描述：角色封禁
请求方式：POST(APPLICATION/JSON)
请求参数：
|字段|类型|是否必填|说明
|gameId|long|是|游戏Id
|zoneId|long|是

响应状态：
|字段|类型|说明
|code|int|状态码 0-成功|

请求示例：
http://127.0.0.1:8080/api/v2/kill/role.do
```json
{
    "gameId": 114,
    "zoneId": 108,
    "roleIds": "1,2,3",
    "sign": "xxxxxxxxx",
    "time": 3600,
    "type": 1
}
```
响应示例：
```json
{
  "code": 0,
  "message": "ok",
  "data": [{
    "roleId": 1,
    "error": "找不到角色"
  },{
    "roleId": 1,
    "roleName": "测试名称",
    "error": "",
    "serverId": 1,
    "tunnelId": 123,
    "killTime": 1746707899
  }]
}
```