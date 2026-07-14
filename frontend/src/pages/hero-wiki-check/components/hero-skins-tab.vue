/**
 * 武将皮肤标签页
 *
 * 展示皮肤列表，每个皮肤含基础信息、资源、Spine、收藏册、台词
 */
<script setup lang="ts">
import {HeroCompleteData} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/herowiki_def/models";

defineProps<{
  heroWiki: HeroCompleteData
}>()
</script>

<template>
  <n-collapse>
    <n-collapse-item
        v-for="(skin, index) in heroWiki.Skins"
        :key="index"
        :title="`${index + 1}. ${skin?.Name || '未命名皮肤'} 品质:${skin?.RailType}`"
        :name="index"
    >
      <n-grid :cols="24" :x-gap="16">
        <!-- 皮肤基础信息 -->
        <n-gi :span="24">
          <n-card size="small" :bordered="false">
            <n-descriptions label-placement="left" :column="2" bordered>
              <n-descriptions-item label="皮肤ID">
                <n-tag :bordered="false">{{ skin?.ItemId }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="皮肤拼音">
                <n-tag :bordered="false">{{ skin?.SkinPinYin }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="皮肤类型">
                <n-tag :bordered="false" type="primary">{{ skin?.SkinType }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="获取方式">
                <n-tag :bordered="false" type="success">{{ skin?.GetWay || '未知' }}</n-tag>
              </n-descriptions-item>
              <n-descriptions-item label="画师">
                {{ skin?.OriginalArtDesigner || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="武将座位图-特殊形状">
                {{ skin?.SeatSpecialImg || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="座位破框图">
                {{ skin?.HeroUIExtraIcons || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="台词配音">
                {{ skin?.LinesDubbed || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="原画绘制人">
                {{ skin?.OriginalArtDesigner || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="皮肤所属收藏">
                {{ skin?.CollectionType || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="收藏皮肤详情界面标签图片">
                {{ skin?.CollectionTagImg || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="武将皮肤Spine击杀表演时间(毫秒)">
                {{ skin?.KillShowTime || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="皮肤偏移配置">
                {{ skin?.BodyOffset || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="全身像是否有出框图">
                {{ skin?.HasOutFrameIcon || '未知' }}
              </n-descriptions-item>
              <n-descriptions-item label="状态">
                <n-badge :type="skin?.IsOpen ? 'success' : 'error'"
                         :value="skin?.IsOpen ? '已开放' : '未开放'"/>
              </n-descriptions-item>
              <n-descriptions-item label="开放日期">
                {{ skin?.OpenDate || '无' }}
              </n-descriptions-item>
            </n-descriptions>
          </n-card>
        </n-gi>

        <!-- 皮肤资源 -->
        <n-gi :span="12" v-if="skin?.Resource">
          <n-card size="small" title="资源路径" :bordered="false">
            <n-text depth="3" style="word-break: break-all;">
              {{ skin?.Resource.Path }}
            </n-text>
          </n-card>
        </n-gi>

        <!-- 皮肤Spine信息 -->
        <n-gi :span="12" v-if="skin?.Spine">
          <n-card size="small" title="Spine配置" :bordered="false">
            <n-space vertical>
              <n-space>
                <n-tag v-if="skin?.Spine.IsHasSeatSpine" type="success">座位Spine</n-tag>
                <n-tag v-if="skin?.Spine.IsHasBookSpine" type="success">大立绘Spine</n-tag>
                <n-tag v-if="skin?.Spine.IsHasMainBgSpine" type="success">主界面Spine</n-tag>
                <n-tag v-if="skin?.Spine.IsHasSeatKillSpine" type="warning">击杀Spine</n-tag>
                <n-tag v-if="skin?.Spine.IsHasCollectionBgSpine" type="warning">收藏册Spine</n-tag>
                <n-tag v-if="skin?.Spine.IsHasCollectionCardBgSpineDuplicate" type="warning">武将详情Spine
                </n-tag>
              </n-space>
              <n-text v-if="skin?.Spine.KillFxId">击杀动效: {{ skin?.Spine.KillFxId }}</n-text>
              <n-text v-if="skin?.Spine.KillAudio">击杀音效: {{ skin?.Spine?.KillAudio }}</n-text>
              <n-text v-if="skin?.Spine.SpineAnimAudio">Spine动画音效: {{ skin?.Spine.SpineAnimAudio }}</n-text>
            </n-space>
          </n-card>
        </n-gi>

        <!-- 收藏册信息 -->
        <n-gi :span="24" v-if="skin?.Collection">
          <n-card size="small" title="收藏册信息" :bordered="false">
            <n-space vertical>
              <n-space align="center">
                <n-tag size="large" :bordered="false" type="info">{{ skin?.Collection.Name }}</n-tag>
                <n-text v-if="skin?.Collection.NameImg">图片: {{ skin?.Collection.NameImg }}</n-text>
              </n-space>
              <n-text v-if="skin?.Collection.Desc" depth="2">{{ skin?.Collection.Desc }}</n-text>
              <n-text v-if="skin?.Collection.Weight" depth="3">权重: {{ skin?.Collection.Weight }}</n-text>
            </n-space>
          </n-card>
        </n-gi>

        <!-- 皮肤台词 -->
        <n-gi :span="24" v-if="skin?.Lines?.length">
          <n-card size="small" :title="`皮肤台词`" :bordered="false">
            <n-list>
              <n-list-item v-for="(line, idx) in skin?.Lines" :key="idx">
                <n-thing>
                  <template #header>
                    {{ line?.TabName + `(${line?.Id})` || '台词' }} ({{ line?.Type }})
                  </template>
                  <template #description>
                    <n-text depth="3">{{ line?.Text }}</n-text>
                  </template>
                  <template #footer>
                    <n-tag v-if="line?.AudioId" size="small" :bordered="false">
                      音效: {{ line?.AudioId }}
                    </n-tag>
                    <n-tag v-if="line?.GroupId" size="small" :bordered="false">
                      台词分组ID: {{ line?.GroupId }}
                    </n-tag>
                  </template>
                </n-thing>
              </n-list-item>
            </n-list>
          </n-card>
        </n-gi>
      </n-grid>
    </n-collapse-item>
  </n-collapse>
</template>
