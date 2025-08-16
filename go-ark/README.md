# 当前项目借鉴[no-ark项目](https://gitee.com/xiaoe/noark3)

## 依赖注入
### by tag
### by interface
### struct
```golang
package demo
type UserDao struct {
	
}
type UserService struct {
    // 依赖注入
    UserDao *UserDao `inject:""`
}
```
### interface
```golang
package demo
type UserDao interface {
    GetUserName() string
}
type UserDaoImpl struct {
	UserDao UserDao `inject:""`
}
```
### func
```golang
package demo
type UserGetter func(id int) any
type UserService struct {
    // 依赖注入
    UserGetter UserGetter `inject:""`
}
```
### 共享变量在包下定义
