/**
 * 重放功能调试测试
 *
 * 运行方式：
 * 1. 启动应用: wails3 dev
 * 2. 运行测试: cd frontend && npx playwright test debug-replay.spec.ts
 */

import { test, expect } from '../shared/fixtures';
import { ProtoTestPage } from '../shared/pages/ProtoTestPage';

test.describe('重放功能调试', () => {
  let page: ProtoTestPage;

  test.beforeEach(async ({ page: p }) => {
    page = new ProtoTestPage(p);
    await page.goto();
  });

  test('调试重放结果数据', async ({ page: p }) => {
    console.log('🧪 开始调试重放功能...');

    // 1. 切换到发包改包页签
    await page.clickTabPacket();
    console.log('✅ 已切换到发包改包页签');

    // 2. 检查页面状态
    const hasData = await p.evaluate(() => {
      const messages = (window as any).__recordData?.messages;
      return messages && messages.length > 0;
    });

    console.log(`📊 当前是否有数据: ${hasData}`);

    if (hasData) {
      console.log('✅ 有数据，可以测试重放功能');

      // 3. 点击开始重放
      await page.startReplayButton.click();
      console.log('✅ 已点击开始重放按钮');

      // 4. 等待重放完成
      await p.waitForTimeout(3000);

      // 5. 检查是否切换到重放结果页签
      const currentTabText = await p.evaluate(() => {
        const tabs = Array.from(document.querySelectorAll('div'));
        return tabs.find(el => el.textContent?.includes('重放结果'))?.textContent || 'unknown';
      });
      console.log(`📍 当前页签: ${currentTabText}`);

      // 6. 检查重放结果数据
      const replayData = await p.evaluate(() => {
        return {
          replayResults: (window as any).__replayResults,
          currentReplayResultId: (window as any).__currentReplayResultId,
          currentResult: (window as any).__currentResult,
          activeTab: (window as any).__activeTab
        };
      });

      console.log('📊 重放结果数据:', JSON.stringify(replayData, null, 2));

      // 7. 检查表格数据
      const tableRowCount = await page.messageTable.locator('tbody tr').count();
      console.log(`📋 表格行数: ${tableRowCount}`);

      if (tableRowCount > 0) {
        console.log('✅ 重放结果页签有数据！');
      } else {
        console.log('❌ 重放结果页签没有数据');
      }

    } else {
      console.log('⚠️ 没有数据，请先录制一些数据');
    }

    // 8. 最终调试信息
    const finalDebugInfo = await p.evaluate(() => {
      return {
        url: window.location.href,
        title: document.title,
        bodyHTML: document.body.innerHTML.substring(0, 500)
      };
    });

    console.log('🔍 最终调试信息:', JSON.stringify(finalDebugInfo, null, 2));
  });

  test('在页面中显示调试面板', async ({ page: p }) => {
    console.log('🧪 在页面中注入调试面板...');

    // 注入调试面板
    await p.evaluate(() => {
      // 创建调试面板
      const panel = document.createElement('div');
      panel.id = 'debug-panel';
      panel.style.cssText = `
        position: fixed;
        top: 10px;
        right: 10px;
        width: 350px;
        max-height: 500px;
        overflow-y: auto;
        background: rgba(0, 0, 0, 0.95);
        color: #0f0;
        padding: 15px;
        border-radius: 8px;
        font-family: 'Courier New', monospace;
        font-size: 12px;
        z-index: 999999;
        border: 2px solid #0f0;
        box-shadow: 0 4px 12px rgba(0,0,0,0.5);
      `;

      panel.innerHTML = `
        <div style="margin-bottom: 10px; font-weight: bold; font-size: 14px; color: #0ff;">
          🐛 调试面板 - 实时监控
        </div>
        <div id="debug-content" style="line-height: 1.6;"></div>
      `;

      document.body.appendChild(panel);

      // 更新调试信息的函数
      (window as any).__updateDebugPanel = function() {
        const content = document.getElementById('debug-content');
        if (!content) return;

        const debugInfo = {
          timestamp: new Date().toLocaleTimeString(),
          replayResults: (window as any).__replayResults || [],
          currentReplayResultId: (window as any).__currentReplayResultId || null,
          activeTab: (window as any).__activeTab || 'unknown',
          events: ((window as any).__debugEvents || []).slice(-5)
        };

        content.innerHTML = `
          <div style="border-bottom: 1px dashed #0f0; padding-bottom: 8px; margin-bottom: 8px;">
            <strong style="color: #0ff;">时间:</strong> ${debugInfo.timestamp}
          </div>

          <div style="border-bottom: 1px dashed #0f0; padding-bottom: 8px; margin-bottom: 8px;">
            <strong style="color: #0ff;">当前页签:</strong> ${debugInfo.activeTab}
          </div>

          <div style="border-bottom: 1px dashed #0f0; padding-bottom: 8px; margin-bottom: 8px;">
            <strong style="color: #0ff;">当前重放ID:</strong> ${debugInfo.currentReplayResultId || '无'}
          </div>

          <div style="border-bottom: 1px dashed #0f0; padding-bottom: 8px; margin-bottom: 8px;">
            <strong style="color: #0ff;">重放结果数:</strong> ${debugInfo.replayResults.length}
          </div>

          <div style="border-bottom: 1px dashed #0f0; padding-bottom: 8px; margin-bottom: 8px;">
            <strong style="color: #0ff;">最近结果:</strong><br>
            ${debugInfo.replayResults.slice(-1).map(r => `  - ID: ${r.id} | 来源: ${r.source} | 消息数: ${r.recordData.message_count} | 状态: ${r.status}`).join('<br>') || '  无结果'}
          </div>

          <div>
            <strong style="color: #0ff;">调试事件:</strong><br>
            ${debugInfo.events.map(e => `  [${e.timestamp}] ${e.message}`).join('<br>') || '  无事件'}
          </div>
        `;
      };

      // 监听页面事件并记录
      (window as any).__debugEvents = [];
      const originalEmit = (window as any).__wails_events;
      if (originalEmit) {
        // 这里可以监听 Wails 事件
        (window as any).__debugEvents.push({
          timestamp: new Date().toLocaleTimeString(),
          message: '调试面板已初始化'
        });
      }

      // 定期更新调试面板
      setInterval((window as any).__updateDebugPanel, 1000);
    });

    console.log('✅ 调试面板已注入，每秒自动更新');

    // 保持测试运行，让用户可以观察调试面板
    await p.waitForTimeout(10000);

    // 获取最终的调试信息
    const finalInfo = await p.evaluate(() => {
      return {
        replayResults: (window as any).__replayResults,
        currentReplayResultId: (window as any).__currentReplayResultId,
        debugEvents: (window as any).__debugEvents
      };
    });

    console.log('📊 最终调试信息:', JSON.stringify(finalInfo, null, 2));
  });
});