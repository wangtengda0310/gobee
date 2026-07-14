// bittorrent-dht 无法在浏览器中运行（依赖 Node.js UDP socket）
// WebTorrent 浏览器版不使用 DHT，此 shim 让 webtorrent 判定 DHT 不可用
const DHTShim = {
    Client: null
}
export default DHTShim
export const Client = null
