/**
 * 语音资源检查页测试
 * 对应 src/pages/hero-voice-resource-check/index.vue
 */

import { test, expect, describe } from '../shared/fixtures';
import { HeroVoiceResourceCheckPage } from '../shared/pages/HeroVoiceResourceCheckPage';

describe('语音资源检查页测试', () => {
  let voicePage: HeroVoiceResourceCheckPage;

  test.beforeEach(async () => {
    voicePage = new HeroVoiceResourceCheckPage(test.getPage());
    await voicePage.goto();
  });

  describe('路径配置', () => {
    test('页面加载 - 显示配置区域', async () => {
      await voicePage.expectConfigAreaVisible();
    });

    test('配表路径配置 - 输入配表位置', async () => {
      const testPath = '/test/excel/path';
      await voicePage.setExcelDir(testPath);
    });

    test('Card文件夹配置 - 输入Card文件夹路径', async () => {
      const testPath = '/test/card/path';
      await voicePage.setCardDir(testPath);
    });
  });

  describe('执行检索', () => {
    // 需要后端支持
    test.skip('开始检索 - 点击开始检索按钮', async () => {
      await voicePage.clickStartCheck();
      await voicePage.waitForCheckComplete();
    });

    test.skip('加载状态 - 检索中显示loading', async () => {
      await voicePage.clickStartCheck();
      await voicePage.expectIsLoading();
      await voicePage.waitForCheckComplete();
      await voicePage.expectNotLoading();
    });
  });

  describe('错误列表', () => {
    test.skip('错误列表显示 - 显示检查错误', async () => {
      // 需要先执行检索
      const count = await voicePage.getErrorCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });

    test.skip('武将分组 - 按武将分组显示错误', async () => {
      const heroGroups = await voicePage.getHeroGroups();
      expect(heroGroups.length).toBeGreaterThanOrEqual(0);
    });

    test.skip('错误详情 - 显示音频错误详情', async () => {
      const errorItem = voicePage.getErrorItem(0);
      await voicePage.expectVisible(errorItem);
    });

    test.skip('音频ID显示 - 显示音频ID', async () => {
      const audioId = await voicePage.getAudioId(0);
      expect(audioId).toBeTruthy();
    });

    test.skip('重复使用次数 - 显示重复使用次数', async () => {
      const repeatNum = await voicePage.getRepeatNum(0);
      expect(repeatNum).toBeGreaterThanOrEqual(1);
    });
  });

  describe('页面布局验证', () => {
    test('页面布局 - 验证配置区域显示', async () => {
      await voicePage.expectConfigAreaVisible();
    });

    test('空结果提示 - 无错误时显示提示', async () => {
      // 未执行检查时，应无错误列表
      await voicePage.expectNoErrors();
    });

    test('按钮状态 - 验证检索按钮', async () => {
      await voicePage.expectCheckButtonEnabled();
    });
  });

  describe('集成测试', () => {
    test.skip('完整流程 - 配置并执行检查', async () => {
      // 配置路径
      await voicePage.setExcelDir('/path/to/excel');
      await voicePage.setCardDir('/path/to/card');

      // 执行检查
      await voicePage.clickStartCheck();
      await voicePage.waitForCheckComplete();

      // 验证结果
      const count = await voicePage.getErrorCount();
      expect(count).toBeGreaterThanOrEqual(0);
    });
  });
});
