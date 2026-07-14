# Hero Voice Resource Check - 武将语音资源检查页面

> File path: `src/pages/hero-voice-resource-check/index.vue`
> Route: `/HeroRes`

## Overview

武将语音资源检查页面，用于检查武将语音资源是否存在及重复使用情况。采用简单的表单 + 列表布局。

## ASCII Layout Diagram

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ ┌────────────────────────────────────────────────────────────────────────┐  │
│ │ 配表位置: [____________________________]                                │  │
│ │ Card文件夹位置: [_________________________]                              │  │
│ │                                                          [开始检索]     │  │
│ └────────────────────────────────────────────────────────────────────────┘  │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐    │
│  │ 武将: 10001(张飞)                                                    │    │
│  │ ┌──────────────────────────────────────────────────────────────────┐ │    │
│  │ │ 音频Id: hero_line_10001_1                                      │ │    │
│  │ │ 重复使用次数: 2                                                  │ │    │
│  │ │ 使用位置: 10001, 10002                                           │ │    │
│  │ │ 原因: 文件不存在, 文件不存在                                     │ │    │
│  │ └──────────────────────────────────────────────────────────────────┘ │    │
│  └──────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│  ┌──────────────────────────────────────────────────────────────────────┐    │
│  │ 武将: 10002(关羽)                                                    │    │
│  │ ┌──────────────────────────────────────────────────────────────────┐ │    │
│  │ │ 音频Id: hero_line_10002_1                                      │ │    │
│  │ │ 重复使用次数: 3                                                  │ │    │
│  │ │ 使用位置: 10002, 10003, 10004                                   │ │    │
│  │ │ 原因: 文件不存在, 文件不存在, 文件不存在                         │ │    │
│  │ └──────────────────────────────────────────────────────────────────┘ │    │
│  └──────────────────────────────────────────────────────────────────────┘    │
│                                                                              │
│                              （可滚动）                                      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Component Tree Structure

```
pages/hero-voice-resource-check/index.vue          # Main page
└── n-scrollbar → index.vue:84
    ├── .path-config-row → index.vue:85
    │   ├── PathConfigInput → @shared/components/path-config-input
    │   │   ├── n-input-group (配表位置)
    │   │   └── n-input-group (Card文件夹位置)
    │   └── n-button (开始检索) → index.vue:95
    │       └── @click="startReport"
    └── .err-list (v-if="testLog") → index.vue:98
        ├── HeroLineVoiceRepeatVoiceMap (v-for) → index.vue:99
        │   └── .errInfo (条件渲染: 文件不存在) → index.vue:101
        │       ├── 武将名称
        │       ├── 音频Id
        │       ├── 重复使用次数
        │       ├── 使用位置
        │       └── 原因
        └── HeroAudioRepeatVoiceMap (v-for) → index.vue:125
            └── .errInfo (条件渲染: 文件不存在) → index.vue:127
                ├── 武将名称
                ├── 重复使用次数
                ├── 使用位置
                └── 原因
```

## Layout Areas

| Area | Size | Description |
|------|------|-------------|
| Path Config Row | Auto | 路径配置表单 |
| Error List | Adaptive | 错误信息列表（可滚动） |

## Component File Mapping

| Component | File Path | Line | Description |
|-----------|-----------|------|-------------|
| Main Page | pages/hero-voice-resource-check/index.vue | 83-136 | n-scrollbar container |
| Path Config | pages/hero-voice-resource-check/index.vue | 85-96 | PathConfigInput + Button |
| Error List | pages/hero-voice-resource-check/index.vue | 98-135 | v-for error items |

## Key State

| State | File | Type | Description |
|-------|------|------|-------------|
| `cardDir` | pages/hero-voice-resource-check/index.vue | Ref\<string\> | Card 文件夹路径 |
| `excelDir` | pages/hero-voice-resource-check/index.vue | Ref\<string\> | Excel 配置目录路径 |
| `testLog` | pages/hero-voice-resource-check/index.vue | Ref\<VoiceCheckReport\> | 检查报告 |
| `errHeroLines` | pages/hero-voice-resource-check/index.vue | ComputedRef | 错误的台词语音 |
| `excelHeroMap` | @config/Hero.ts | Map | 武将配置映射 |

## Data Flow

```
User Action
    │
    ▼
Click "开始检索"
    │
    ▼
startReport()
    │
    ├──► 构建音频目录路径 (cardDir + "/Audio/")
    ├──► HeroResCheckService.Check(excelDir, audioDir)
    │         │
    │         ▼
    │    testLog.value = res
    │         │
    │         ▼
    └────► errHeroLines computed update
              │
              └──► v-for render error list
```

## Related Files

### Config
- `@config/Hero.ts` - `excelHeroMap` - 武将配置映射

### Backend Service
- `@bindings/rain-qa-func/internal` - `HeroResCheckService.Check()` - 语音检查服务

## Data Structures

### VoiceCheckReport
```typescript
interface VoiceCheckReport {
  HeroLineVoiceRepeatVoiceMap: {
    [heroId: string]: {
      [audioId: string]: {
        RepeatNum: number
        Location: string[]
        ExistInfos: { Reason: string; Exist: boolean }[]
      }
    }
  }
  HeroAudioRepeatVoiceMap: {
    [heroId: string]: {
      [audioId: string]: {
        RepeatNum: number
        Location: string[]
        ExistInfos: { Reason: string; Exist: boolean }[]
      }
    }
  }
}
```

## Event Flow

```
Component          Event                  Handler                  Effect
────────────────────────────────────────────────────────────────────────────
n-button           @click                 → startReport()          执行检查
PathConfigInput    @update:excel-dir       → excelDir update        更新配表路径
PathConfigInput    @update:second-value    → cardDir update         更新 Card 路径
```

## Notes

- 使用 PathConfigInput 组件配置路径（layout="flex" 模式）
- 检查两种类型：HeroLineVoiceRepeatVoiceMap（台词语音）、HeroAudioRepeatVoiceMap（武将语音）
- 只显示存在错误的项（ExistInfos 中存在 Exist=false 的项）
- 错误信息包含：武将名称、音频 ID、重复使用次数、使用位置、原因
- 使用 `@config/Hero.ts` 中的 `excelHeroMap` 获取武将名称
- 列表按武将分组显示
