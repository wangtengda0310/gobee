import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import wails from "@wailsio/runtime/plugins/vite";

// https://vitejs.dev/config/
// 注意: wails 插件参数指向 bindings 子目录 (generate bindings -d frontend/bindings 的产物)
export default defineConfig({
  server: {
    host: "127.0.0.1",
    port: Number(process.env.WAILS_VITE_PORT) || 9246,
    strictPort: true,
  },
  plugins: [vue(), wails("./bindings")],
});
