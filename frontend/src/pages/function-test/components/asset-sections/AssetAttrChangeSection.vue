<!-- AttrChange 类型资产段 — 出杀次数/手牌上限/装备上限/攻击范围/距离/生命 -->
<script setup lang="ts">
const props = defineProps<{
  assetList: { [key: string]: any }
}>()

const emit = defineEmits<{
  (e: 'update'): void
}>()
</script>

<template>
  <div class="attrChangeTypeAsset">
    <div>
      <div>出杀次数</div>
      <n-input-number v-model:value="props.assetList.ShaCount" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="出杀次数"
                      clearable/>
    </div>
    <div>
      <div>手牌上限</div>
      <n-input-number v-model:value="props.assetList.HandLimit" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="手牌上限"
                      clearable/>
    </div>
    <div>
      <div>装备上限</div>
      <n-input-number v-model:value="props.assetList.EquipLimit" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="装备上限"
                      clearable/>
    </div>
    <div>
      <div>攻击范围</div>
      <n-input-number v-model:value="props.assetList.AttackRange" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="攻击范围"
                      clearable/>
    </div>
    <div>
      <div>距离增加</div>
      <n-input-number v-model:value="props.assetList.DistIncr" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="距离增加"
                      clearable/>
    </div>
    <div>
      <div>距离减少</div>
      <n-input-number v-model:value="props.assetList.DistDecr" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="距离减少"
                      clearable/>
    </div>
    <div>
      <div>最大生命</div>
      <n-input-number v-model:value="props.assetList.MaxHp" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="最大生命"
                      clearable/>
    </div>
    <div>
      <div>当前生命</div>
      <n-input-number v-model:value="props.assetList.CurHp" min="-999999" max="999999"
                      @update:value="emit('update')" placeholder="当前生命"
                      clearable/>
    </div>
    <n-switch v-model:value="props.assetList.UnExpect"
              @update:value="emit('update')"
              :round="false">
      <template #checked>
        不期望
      </template>
      <template #unchecked>
        期望
      </template>
    </n-switch>
  </div>
</template>
