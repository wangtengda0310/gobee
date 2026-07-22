# pcapdump：离线 .pcap 文件解析示例

用 pcap 包读取已有的 `.pcap` 文件，逐包解析并输出各层（Ethernet/IPv4/TCP/UDP/HTTP 等）的结构化信息。

**纯 Go 程序**——不需要 Npcap / cgo，任意平台可运行。

## 用法

```bash
cd pcap

# 基本用法：打印每个包的各层摘要
go run ./examples/pcapdump -file dump.pcap

# 只显示 TCP 包
go run ./examples/pcapdump -file dump.pcap -filter tcp

# 显示应用层数据（HTTP 请求行等明文内容）
go run ./examples/pcapdump -file dump.pcap -app

# 统计模式：只输出协议分布统计
go run ./examples/pcapdump -file dump.pcap -stats
```

## 参数

| 参数 | 默认值 | 说明 |
|---|---|---|
| `-file` | （必填） | pcap 文件路径 |
| `-filter` | 空 | 只显示指定协议（tcp/udp/arp/icmp/ipv4） |
| `-app` | false | 显示应用层数据第一行（HTTP 请求行等） |
| `-stats` | false | 只输出统计，不逐包打印 |

## 获取测试用的 pcap 文件

### 方法 1：用 pcaptest 抓取（需 Npcap）

```bash
CGO_ENABLED=1 go run -tags livecapture ./cmd/pcaptest \
  -iface "\Device\NPF_{...}" -out dump.pcap
# Ctrl+C 停止后，dump.pcap 包含抓到的所有包
```

### 方法 2：用 Wireshark / tcpdump

```bash
# tcpdump
sudo tcpdump -i eth0 -w dump.pcap tcp port 80

# Wireshark：菜单 → File → Save As → .pcap 格式
```

### 方法 3：用 buf demo 的 sniffer 生成

（sniffer 当前只做实时解析，不写文件——但你可以用 pcaptest 的 `-out` 参数）

## 输出示例

### 逐包模式（默认）

```
文件: dump.pcap
链路层: Ethernet
SnapLen: 65535

#0001 [10:00:00.123456] len=74  ETH 00:11:..→66:77:..  IPv4 10.0.0.2→10.0.0.3  proto=TCP  TCP 51000→80 SYN
#0002 [10:00:00.123789] len=66  ETH 00:11:..→66:77:..  IPv4 10.0.0.3→10.0.0.2  proto=TCP  TCP 80→51000 SYN/ACK
#0003 [10:00:00.124001] len=66  ETH 00:11:..→66:77:..  IPv4 10.0.0.2→10.0.0.3  proto=TCP  TCP 51000→80 ACK
#0004 [10:00:00.125000] len=170 ETH 00:11:..→66:77:..  IPv4 10.0.0.2→10.0.0.3  proto=TCP  TCP 51000→80 PSH/ACK
      app: GET / HTTP/1.1
#0005 [10:00:00.126000] len=340 ETH 00:11:..→66:77:..  IPv4 10.0.0.3→10.0.0.2  proto=TCP  TCP 80→51000 PSH/ACK
      app: HTTP/1.1 200 OK

=== 统计 ===
总包数: 5
  Ethernet: 5
  IPv4: 5, IPv6: 0
  TCP: 5, UDP: 0
  ARP: 0, ICMP: 0
  应用层数据: 2
```

### 统计模式（`-stats`）

```
文件: dump.pcap
链路层: Ethernet
SnapLen: 65535

=== 统计 ===
总包数: 1247
  Ethernet: 1247
  IPv4: 1200, IPv6: 0
  TCP: 980, UDP: 220
  ARP: 12, ICMP: 0
  应用层数据: 456
```

## 技术要点

- 使用 `pcap.NewReaderSource` 把 `pcapgo.Reader` 适配为 pcap 包的 `Source` 接口
- 直接 `range src.Packets()` 遍历（离线场景不需要 Capturer 的广播/过载保护）
- 用 gopacket 的 `pkt.Layer(LayerTypeXxx)` 逐层提取信息
- 解析错误不致命（`pkt.ErrorLayer()` 检测），不影响后续包
