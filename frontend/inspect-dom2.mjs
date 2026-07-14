import { chromium } from 'playwright';

const browser = await chromium.connectOverCDP('http://localhost:9223');
const contexts = browser.contexts();
const page = contexts[0].pages()[0];

await page.waitForTimeout(1000);

// 清理模态框
await page.keyboard.press('Escape');
await page.waitForTimeout(300);

// 导航到协议重放页
await page.locator('#layout-header button:has-text("Proto测试")').click();
await page.waitForTimeout(1000);

console.log('=== 当前页签检查 ===');
const tabs = await page.locator('button:has-text("发包改包"), button:has-text("测试用例")').all();
for (let i = 0; i < tabs.length; i++) {
  const text = await tabs[i].textContent();
  const visible = await tabs[i].isVisible();
  console.log(`Tab ${i}: "${text?.trim()}" visible=${visible}`);
}

// 查看所有按钮（包括可能隐藏的）
const allButtons = await page.locator('button', { hasText: '录制' }).or(page.locator('button', { hasText: '重放' })).or(page.locator('button', { hasText: '多选' })).all();
console.log('\n=== 录制/重放/多选按钮 ===');
for (let i = 0; i < allButtons.length; i++) {
  const text = await allButtons[i].textContent();
  const visible = await allButtons[i].isVisible();
  console.log(`${i}: "${text?.trim()}" visible=${visible}`);
}

// 切换到发包改包页签（如果不在）
await page.locator('button:has-text("发包改包")').first().click();
await page.waitForTimeout(500);

console.log('\n=== 发包改包页签按钮 ===');
const packetButtons = await page.locator('button').all();
for (let i = 0; i < packetButtons.length; i++) {
  const text = await packetButtons[i].textContent();
  const visible = await packetButtons[i].isVisible();
  if (visible) {
    console.log(`${i}: "${text?.trim()}"`);
  }
}

await browser.close();
