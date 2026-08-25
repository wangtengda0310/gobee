# webrtc-boids — Go 端 WebRTC 鸟群节点

与浏览器页面 `webrtc-boids.html`（见 itsnot.fun 项目）互联的技术验证程序：
通过公共 MQTT（EMQX）做信令，与 HTML 页面建立 WebRTC DataChannel，
各自仿真自己生成的鸟（实体归属权威），15Hz 快照广播，组成同一个鸟群系统。

## 运行

```bash
go build -o go-boids .

# 默认：EMQX、房间 boids-room-1、flock 角色、8 只鸟
./go-boids

# 捕食者角色、指定数量与位置
./go-boids -role predator -n 3 -cx 200 -cy 200

# 指定 broker / 房间（需与 HTML 页面一致）
./go-boids -broker tcp://test.mosquitto.org:1883 -room my-room
```

HTML 侧：打开 `webrtc-boids.html` → 信令选「公共 MQTT / EMQX」→ 房间名一致 → 连接。
Go 端日志每 2 秒打印：`📊 本地鸟=N · 远端实体=M · 连接 X/Y`。

## 协议（与 HTML 页面对齐）

- MQTT 前缀 `webrtc-boids-v1/<room>`：`/pres` 在线名单（hi/hello/bye + 5s 心跳），`/sig/<peerId>` 定向信令
- 发起方 = peerId 字典序小的一侧（tie-break，无 glare）；offer 一律清场重答；8s 未连自动重试
- DataChannel `boids`（ordered）：`{t:'hello'}` / `{t:'state', e:[{id,k,x,y,vx,vy,d?}]}` / `{t:'clear'}` / `{t:'ping'/'pong'}`
- 快照坐标归一化到 [0,800)×[0,600) 保留 1 位小数；远端插值/环面由接收方处理

## 实测记录（2026-08-25）

- 与 HTML 页面互连：MQTT 发现 → offer/answer → DC 打开 < 2s，双向 state 同步稳定
- 三方 mesh（页面 + 2×Go 进程）：各自看到全部远端实体，连接 2/2
