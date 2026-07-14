/**
 * Buff 显示组件
 *
 * 显示技能关联的 Buff 信息，包括基础信息和详细属性
 */
<script setup lang="ts">
import {computed} from 'vue';
import {NBadge, NDivider, NEmpty, NGi, NGrid, NPopover, NSpace, NTag, NText} from 'naive-ui';
import {BuffInfo} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def";

const props = defineProps<{
  buffs: (BuffInfo | null)[];
}>();

// 过滤掉null值
const validBuffs = computed(() => {
  return props.buffs.filter(buff => buff !== null) as BuffInfo[];
});

// 获取Buff类型样式和文本
const getBuffTypeInfo = (type: number) => {
  const types: Record<number, { type: 'success' | 'error' | 'warning' | 'info' | 'default'; text: string }> = {
    0: {type: 'success', text: '0'},
    1: {type: 'error', text: '1'},
    2: {type: 'warning', text: '2'},
    3: {type: 'info', text: '3'},
    4: {type: 'default', text: '4'},
  };
  return types[type] || {type: 'default', text: '未知'};
};

// 获取叠加类型文本
const getOverlyingTypeText = (type: number) => {
  const types: Record<number, string> = {
    0: '0',
    1: '1',
    2: '2',
    3: '3',
    4: '4',
    5: '5',
    6: '6',
    7: '7',
  };
  return types[type] || type.toString();
};

// 获取结束类型文本
const getEndTypeText = (type: string) => {
  const types: Record<string, string> = {
    '0': '无',
    '1': '执行者回合开始',
    '2': '施法者回合开始',
    '3': '每回合开始',
    '4': '执行者回合结束',
    '5': '施法者回合结束',
    '6': '每回合开始结束',
    '7': '循环开始',
    '8': '执行者出牌阶段结束',
    '9': '执行者抽牌阶段结束',
  };
  return types[type] || type || '未知';
};

// 格式化数组显示
const formatArray = (arr: any[] | undefined, defaultValue: string = '无'): string => {
  if (!arr || arr.length === 0) return defaultValue;
  return arr.join(', ');
};

// 格式化布尔值
const formatBoolean = (value: boolean | undefined, trueText: string = '是', falseText: string = '否'): string => {
  return value ? trueText : falseText;
};

// 获取显示区域文本
const getShowAreaText = (area: string): string => {
  const areas: Record<string, string> = {
    'None': '空Buff，不显示',
    'EBuffShowArea_SeatLeftArea': '座位左边栏Buff',
    'EBuffShowArea_SeatRightArea': '座位右边栏Buff',
    'EBuffShowArea_SeatCustom': '座位自定义Buff(带Prefab)',
    'EBuffShowArea_SeatNoPrefab': '座位自定义Buff(不带Prefab)',
    'EBuffShowArea_Room': '房间Buff（在房间左上Buff栏显示）',
    'EBuffShowArea_RoomCustom': '房间自定义位置Buff（自定义显示位置）',
    'EBuffShowArea_RoomNoPrefab': '房间无Prefab的Buff',
  };
  return areas[area] || area || '默认区域';
};

// 获取所有者类型文本
const getOwnerTypeText = (type: number): string => {
  const types: Record<number, string> = {
    0: '0',
    1: '1',
    2: '2',
    3: '3',
  };
  return types[type] || '未知';
};

// 获取转移类型文本
const getTransferTypeText = (type: number): string => {
  const types: Record<number, string> = {
    0: '转移',
    1: '不转移buff,buff仍生效',
    2: '不转移buff,但buff失效',
  };
  return types[type] || '转移';
};
</script>

<template>
  <div class="buff-display-dark">
    <template v-if="validBuffs.length > 0">
      <n-space vertical :size="8">
        <div
            v-for="buff in validBuffs"
            :key="buff.Id"
            class="buff-item-dark"
        >
          <n-popover trigger="hover" placement="top-start" :width="360">
            <template #trigger>
              <div class="buff-trigger-dark">
                <!-- Buff图标区域 -->
                <div class="buff-icon-dark" :class="`buff-type-${buff.BuffType}`">
                  <span v-if="buff.Icon" class="buff-icon-text">{{ buff.Icon }}</span>
                  <span v-else class="buff-icon-placeholder">
                    {{ buff.BuffType === 1 ? '✨' : buff.BuffType === 2 ? '💢' : '⚡' }}
                  </span>
                </div>

                <!-- Buff基本信息 -->
                <div class="buff-info-dark">
                  <div class="buff-header-dark">
                    <span class="buff-name-dark" :title="buff.Name">
                      {{ buff.Name || `Buff ${buff.Id}` }}
                    </span>
                    <div class="buff-badges-dark">
                      <n-tag
                          v-if="buff.BuffDot === 1"
                          size="tiny"
                          :bordered="false"
                          style="background: #d97706; color: white;"
                      >
                        DOT
                      </n-tag>
                      <n-tag
                          v-if="buff.IsTrigger"
                          size="tiny"
                          :bordered="false"
                          style="background: #059669; color: white;"
                      >
                        触发
                      </n-tag>
                      <n-tag
                          v-if="buff.IsServerOnly"
                          size="tiny"
                          :bordered="false"
                          style="background: #4f46e5; color: white;"
                      >
                        服务端
                      </n-tag>
                    </div>
                  </div>

                  <div class="buff-meta-dark">
                    <span class="buff-type-tag" :class="`buff-type-${buff.BuffType}`">
                      {{ getBuffTypeInfo(buff.BuffType).text }}
                    </span>
                    <span class="buff-duration-dark">
                      <span class="duration-item">{{ buff.Round >= 999 ? '∞' : buff.Round }}回</span>
                      <span class="duration-item">{{ buff.Value >= 999 ? '∞' : buff.Value }}次</span>
                    </span>
                  </div>

                  <div v-if="buff.EffectDescribe" class="buff-desc-dark">
                    {{ buff.EffectDescribe }}
                  </div>
                </div>
              </div>
            </template>

            <!-- Popover详情 - 完整Buff信息 -->
            <div class="buff-detail-popover">
              <div class="popover-header">
                <span class="popover-title">{{ buff.Name || `Buff ${buff.Id}` }}</span>
                <n-badge
                    :type="getBuffTypeInfo(buff.BuffType).type"
                    :value="getBuffTypeInfo(buff.BuffType).text"
                />
              </div>

              <n-divider style="margin: 12px 0;"/>

              <n-grid :cols="2" :x-gap="12" :y-gap="8">
                <!-- 基础信息列 -->
                <n-gi>
                  <div class="popover-section">
                    <div class="section-title">基础信息</div>
                    <div class="info-row">
                      <span class="label">BuffID:</span>
                      <span class="value">{{ buff.Id }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">EnumID:</span>
                      <span class="value">{{ buff.EBuffId || '无' }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">类型:</span>
                      <span class="value">{{ getBuffTypeInfo(buff.BuffType).text }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">优先级:</span>
                      <span class="value">{{ buff.BuffPriority }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">叠加类型:</span>
                      <span class="value">{{ getOverlyingTypeText(buff.OverlyingType) }}</span>
                    </div>
                  </div>
                </n-gi>

                <!-- 持续时间列 -->
                <n-gi>
                  <div class="popover-section">
                    <div class="section-title">持续时间</div>
                    <div class="info-row">
                      <span class="label">生效回合:</span>
                      <span class="value">{{ buff.Round >= 999 ? '∞' : buff.Round }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">生效次数:</span>
                      <span class="value">{{ buff.Value >= 999 ? '∞' : buff.Value }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">结束类型:</span>
                      <span class="value">{{ getEndTypeText(buff.EndType) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">次数耗尽结束:</span>
                      <span class="value">{{ formatBoolean(buff.EndByEffect) }}</span>
                    </div>
                  </div>
                </n-gi>

                <!-- 生死规则 -->
                <n-gi>
                  <div class="popover-section">
                    <div class="section-title">生死规则</div>
                    <div class="info-row">
                      <span class="label">施法者死亡:</span>
                      <span class="value">{{ formatBoolean(buff.IsDeleteByCasterDead, '移除', '保留') }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">执行者死亡:</span>
                      <span class="value">{{ formatBoolean(buff.IsDeleteByExecutorDead, '移除', '保留') }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">封禁后有效:</span>
                      <span class="value">{{ formatBoolean(buff.IsValidByFengJin) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">删除技能保留:</span>
                      <span class="value">{{ formatBoolean(buff.IsReserveByRemoveSkill) }}</span>
                    </div>
                  </div>
                </n-gi>

                <!-- 转移规则 -->
                <n-gi>
                  <div class="popover-section">
                    <div class="section-title">转移规则</div>
                    <div class="info-row">
                      <span class="label">转移类型:</span>
                      <span class="value">{{ getTransferTypeText(buff.TransferSkillBuffType) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">所有者类型:</span>
                      <span class="value">{{ getOwnerTypeText(buff.OwnerType) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">仅施法者可见:</span>
                      <span class="value">{{ formatBoolean(buff.IsCasterOnly) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="label">仅服务端:</span>
                      <span class="value">{{ formatBoolean(buff.IsServerOnly) }}</span>
                    </div>
                  </div>
                </n-gi>
              </n-grid>

              <!-- 显示区域 -->
              <div class="popover-section" v-if="buff.ShowArea">
                <div class="section-title">显示区域</div>
                <div class="info-row">
                  <span class="label">区域:</span>
                  <span class="value">{{ getShowAreaText(buff.ShowArea) }}</span>
                </div>
              </div>

              <!-- 属性修改 -->
              <div class="popover-section" v-if="buff.BuffPro?.length">
                <div class="section-title">属性修改</div>
                <div class="info-row">
                  <span class="label">属性:</span>
                  <span class="value">{{ formatArray(buff.BuffPro) }}</span>
                </div>
                <div class="info-row" v-if="buff.ProValue?.length">
                  <span class="label">数值:</span>
                  <span class="value">{{ formatArray(buff.ProValue) }}</span>
                </div>
                <div class="info-row" v-if="buff.CostEffectValue?.length">
                  <span class="label">消耗生效次数:</span>
                  <span class="value">{{ formatArray(buff.CostEffectValue) }}</span>
                </div>
              </div>

              <!-- 触发条件 -->
              <div class="popover-section" v-if="buff.TriggerCondition?.length || buff.TriggerTiming?.length">
                <div class="section-title">触发条件</div>
                <div class="info-row" v-if="buff.TriggerTiming?.length">
                  <span class="label">触发时机:</span>
                  <span class="value">{{ formatArray(buff.TriggerTiming) }}</span>
                </div>
                <div class="info-row" v-if="buff.TriggerCondition?.length">
                  <span class="label">触发条件:</span>
                  <span class="value">{{ formatArray(buff.TriggerCondition) }}</span>
                </div>
                <div class="info-row" v-if="buff.TriggerPriority?.length">
                  <span class="label">触发优先级:</span>
                  <span class="value">{{ formatArray(buff.TriggerPriority) }}</span>
                </div>
                <div class="info-row" v-if="buff.TriggerAction?.length">
                  <span class="label">触发行为:</span>
                  <span class="value">{{ formatArray(buff.TriggerAction) }}</span>
                </div>
              </div>

              <!-- 状态和特效 -->
              <div class="popover-section" v-if="buff.BuffState?.length">
                <div class="section-title">状态特效</div>
                <div class="info-row">
                  <span class="label">Buff状态:</span>
                  <span class="value">{{ formatArray(buff.BuffState) }}</span>
                </div>
                <div class="info-row" v-if="buff.Effect">
                  <span class="label">Loop特效:</span>
                  <span class="value">{{ buff.Effect }}</span>
                </div>
                <div class="info-row" v-if="buff.FlashEffect">
                  <span class="label">闪动特效:</span>
                  <span class="value">{{ buff.FlashEffect }}</span>
                </div>
              </div>

              <!-- 记录信息 -->
              <div class="popover-section" v-if="buff.NeedRecord">
                <div class="info-row">
                  <span class="label">战报记录:</span>
                  <span class="value">{{ formatBoolean(buff.NeedRecord) }}</span>
                </div>
              </div>
            </div>
          </n-popover>
        </div>
      </n-space>
    </template>

    <n-empty v-else description="无Buff数据" size="small">
      <template #extra>
        <n-text depth="3">该技能没有关联Buff</n-text>
      </template>
    </n-empty>
  </div>
</template>

<style scoped>
.buff-display-dark {
  width: 100%;
}

/* Buff项容器 */
.buff-item-dark {
  width: 100%;
  cursor: pointer;
  transition: all 0.2s ease;
}

/* 触发区域 - 紧凑显示 */
.buff-trigger-dark {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 6px 8px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.buff-trigger-dark:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: rgba(255, 255, 255, 0.1);
  transform: translateX(2px);
}

/* 图标区域 */
.buff-icon-dark {
  width: 32px;
  height: 32px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  flex-shrink: 0;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.buff-type-1 {
  background: rgba(16, 185, 129, 0.15);
  border-color: rgba(16, 185, 129, 0.3);
  color: #10b981;
}

.buff-type-2 {
  background: rgba(239, 68, 68, 0.15);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
}

.buff-type-3 {
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.3);
  color: #f59e0b;
}

.buff-type-4 {
  background: rgba(139, 92, 246, 0.15);
  border-color: rgba(139, 92, 246, 0.3);
  color: #8b5cf6;
}

.buff-icon-placeholder {
  font-size: 18px;
  opacity: 0.8;
}

/* 信息区域 */
.buff-info-dark {
  flex: 1;
  min-width: 0;
}

.buff-header-dark {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  margin-bottom: 2px;
}

.buff-name-dark {
  font-size: 13px;
  font-weight: 500;
  color: #e5e7eb;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.buff-badges-dark {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

:deep(.n-tag--tiny) {
  font-size: 9px;
  padding: 0 4px;
  height: 16px;
}

/* 元信息行 */
.buff-meta-dark {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.buff-type-tag {
  font-size: 10px;
  padding: 0 6px;
  height: 16px;
  line-height: 16px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.05);
  color: #9ca3af;
}

.buff-type-tag.buff-type-1 {
  background: rgba(16, 185, 129, 0.15);
  color: #10b981;
}

.buff-type-tag.buff-type-2 {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}

.buff-type-tag.buff-type-3 {
  background: rgba(245, 158, 11, 0.15);
  color: #f59e0b;
}

.buff-type-tag.buff-type-4 {
  background: rgba(139, 92, 246, 0.15);
  color: #8b5cf6;
}

.buff-duration-dark {
  display: flex;
  gap: 6px;
  font-size: 10px;
  color: #9ca3af;
}

.duration-item {
  position: relative;
}

.duration-item:not(:last-child)::after {
  content: '';
  position: absolute;
  right: -4px;
  top: 3px;
  width: 1px;
  height: 8px;
  background: rgba(255, 255, 255, 0.2);
}

/* 效果描述 */
.buff-desc-dark {
  font-size: 11px;
  color: #9ca3af;
  line-height: 1.5;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Popover详情样式 */
.buff-detail-popover {
  padding: 4px 0;
}

.popover-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.popover-title {
  font-size: 15px;
  font-weight: 600;
  color: #f3f4f6;
}

.popover-section {
  margin-bottom: 12px;
}

.popover-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 12px;
  font-weight: 600;
  color: #d1d5db;
  margin-bottom: 6px;
  padding-left: 4px;
  border-left: 2px solid #4f46e5;
}

.info-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 11px;
  padding: 2px 4px;
  border-radius: 2px;
}

.info-row:hover {
  background: rgba(255, 255, 255, 0.02);
}

.info-row .label {
  min-width: 75px;
  color: #9ca3af;
  font-size: 11px;
}

.info-row .value {
  color: #e5e7eb;
  word-break: break-word;
  flex: 1;
  font-size: 11px;
}

/* 暗黑主题适配 */
:deep(.n-popover) {
  --n-color: #1e1e1e;
  --n-text-color: #e5e7eb;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

:deep(.n-divider) {
  --n-color: rgba(255, 255, 255, 0.1);
}

:deep(.n-empty) {
  --n-text-color: #9ca3af;
}

:deep(.n-tag) {
  --n-color: rgba(255, 255, 255, 0.05);
  --n-text-color: #e5e7eb;
  --n-border: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
