import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { nodePolyfills } from 'vite-plugin-node-polyfills'
import { resolve } from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), nodePolyfills({
    globals: { Buffer: true, global: true, process: true },
  })],
  optimizeDeps: {
    include: [
      'vue',
      'vue-router',
      'naive-ui',
      'element-plus',
      '@element-plus/icons-vue',
      'vue-draggable-plus',
      '@wailsio/runtime',
      '@vicons/antd',
      '@vicons/carbon',
      '@vicons/fa',
      '@vicons/fluent',
      '@vicons/ionicons4',
      '@vicons/ionicons5',
      '@vicons/material',
      '@vicons/tabler',
      'vite-plugin-node-polyfills/shims/buffer',
      'vite-plugin-node-polyfills/shims/global',
      'vite-plugin-node-polyfills/shims/process',
    ],
  },
  server: {
    host: '127.0.0.1',
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
      '@pages': resolve(__dirname, 'src/pages'),
      '@shared': resolve(__dirname, 'src/shared'),
      '@bindings': resolve(__dirname, 'bindings'),
      // WebTorrent: bittorrent-dht 无法在浏览器中运行（依赖 UDP），提供空模块让它跳过 DHT
      'bittorrent-dht': resolve(__dirname, 'src/shared/polyfills/bittorrent-dht.js'),
      'bittorrent-tracker/lib/client/http-tracker.js': resolve(__dirname, 'src/shared/polyfills/empty.js'),
    },
  },
})
