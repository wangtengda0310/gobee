import { chromium } from 'playwright';

const browser = await chromium.connectOverCDP('http://localhost:9223');
const contexts = browser.contexts();
const page = contexts[0].pages()[0];

await page.waitForTimeout(1000);

// 导航到协议重放页
await page.locator('#layout-header button:has-text("Proto测试")').click();
await page.waitForTimeout(1000);

// 查看所有按钮文本
const buttons = await page.locator('button').all();
console.log('=== 按钮总数:', buttons.length, '===');
for (let i = 0; i < buttons.length; i++) {
  const text = await buttons[i].textContent();
  const disabled = await buttons[i].isDisabled();
  console.log(`${i}: "${text?.trim()}" (disabled: ${disabled})`);
}

// 查看输入框 placeholder
const inputs = await page.locator('input').all();
console.log('\n=== 输入框数量:', inputs.length, '===');
for (let i = 0; i < inputs.length; i++) {
  const placeholder = await inputs[i].getAttribute('placeholder');
  const value = await inputs[i].inputValue();
  console.log(`${i}: placeholder="${placeholder}" value="${value}"`);
}

// 切换到测试用例页签
await page.locator('text=测试用例').first().click();
await page.waitForTimeout(500);

// 查看测试用例页签的按钮
const testcaseButtons = await page.locator('button').all();
console.log('\n=== 测试用例页签按钮 ===');
for (let i = 0; i < testcaseButtons.length; i++) {
  const text = await testcaseButtons[i].textContent();
  const disabled = await testcaseButtons[i].isDisabled();
  console.log(`${i}: "${text?.trim()}" (disabled: ${disabled})`);
}

await browser.close();
