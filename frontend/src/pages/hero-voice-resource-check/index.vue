<!--
  HeroVoiceResourceCheck - 武将语音资源检查页面

  功能：
  - 检查武将语音资源是否存在及重复使用情况
  - 显示检查结果：武将名称、音频ID、重复使用次数、使用位置、原因

  依赖：
  - HeroResCheckService: 后端语音检查服务
  - PathConfigInput: 路径配置输入组件
  - excelHeroMap: 武将配置映射
-->
<script setup lang="ts">
import {HeroResCheckService} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker";
import {computed, ref} from "vue";
import {LineVoiceRepeatInfo, VoiceCheckReport} from "@bindings/git.devcloud.ztgame.com/v-tangfangda/rain-qa-func/backend/pkg/rain-resources-checker/hero_res_check";
import {excelHeroMap} from "@shared/config/hero";
import PathConfigInput from "@shared/components/path-config-input/index.vue";

const cardDir = ref("../../client/Master/Card/")
const excelDir = ref("../../config/excel")
const testLog = ref<VoiceCheckReport>()

const startReport = () => {
  testLog.value = {} as any
  let audioDir = ""
  if (cardDir.value.endsWith("/")) {
    audioDir = cardDir.value + "Audio/"
  } else {
    audioDir = cardDir.value + "/Audio/"
  }
  HeroResCheckService.Check(excelDir.value, audioDir).then(res => {
    console.log(res)
    if (res) testLog.value = res
  }).catch(err => {
    console.log(err)
  })
}

const errHeroLines = computed(() => {
  if (!testLog.value || !testLog.value.HeroLineVoiceRepeatVoiceMap) return
  return Object.keys(testLog.value.HeroLineVoiceRepeatVoiceMap).flatMap(k => {
    // 过滤一下含有错误的英雄
    const v = testLog.value?.HeroLineVoiceRepeatVoiceMap[k] as {
      [p: `${number}`]: LineVoiceRepeatInfo | null
    } | undefined
    if (v) {
      return Object.keys(v).filter(_k => {
        return v[_k].ExistInfos.find(ei => !ei.Exist)
      }).map(_k => {
        return [
          k,
          {
            repeatNum: v[_k].RepeatNum,
            location: v[_k].Location,
            reason: v[_k].ExistInfos
          }
        ]
      })
    }
    return []
    // 按英雄分组
  }).filter(l => l.length > 0).reduce((previousValue, currentValue, currentIndex, array) => {
    if (previousValue[currentValue[0] as string]) {
      previousValue[currentValue[0] as string].push(currentValue[1] as {
        repeatNum: any
        location: any
        reason: any
      })
    } else {
      previousValue[currentValue[0] as string] = [currentValue[1] as {
        repeatNum: any
        location: any
        reason: any
      }]
    }
    return previousValue
  }, {})
})
</script>

<template>
  <div id="ResCheck">
    <n-scrollbar>
      <div style="display: flex; align-items: center; padding: 10px; gap: 10px">
        <div style="display: flex; align-items: center; width: 1000px; gap: 10px">
          <PathConfigInput
            v-model:excel-dir="excelDir"
            v-model:second-value="cardDir"
            excel-label="配表位置"
            second-label="Card文件夹位置"
            layout="flex"
          />
        </div>
        <n-button @click="startReport">开始检索</n-button>
      </div>
      <!--      <div>{{ errHeroLines }}</div>-->
      <div v-if="testLog">
        <div v-for="(v, k) in testLog?.HeroLineVoiceRepeatVoiceMap">
          <div v-for="(_v, _k) in v">
            <div v-if="_v?.ExistInfos.find(ei=>!ei.Exist)" class="errInfo">
              <div>
                <span>武将:</span>
                <span>{{ k }}({{ excelHeroMap[k]?.Name }})</span>
              </div>
              <div>
                <span>音频Id:</span>
                <span>{{ _k }}</span>
              </div>
              <div>
                <span>重复使用次数:</span>
                <span>{{ _v?.RepeatNum }}</span>
              </div>
              <div>
                <span>使用位置:</span>
                <span v-for="(loc) in _v?.Location">{{ loc }}</span>
              </div>
              <div>
                <span>原因:</span>
                <span v-for="(info) in _v?.ExistInfos">{{ info.Reason }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-for="(v, k) in testLog.HeroAudioRepeatVoiceMap">
          <div v-for="(_v, _k) in v">
            <div v-if="_v?.ExistInfos.find(ei=>!ei.Exist)">
              <div>{{ k }}({{ excelHeroMap[k]?.Name }}):</div>
              <div>重复使用次数:{{ _v?.RepeatNum }}</div>
              <div>使用位置:<span v-for="(loc) in _v?.Location">{{ loc }}</span></div>
              <div>原因:<span v-for="(info) in _v?.ExistInfos">{{ info.Reason }}</span></div>
            </div>
          </div>
        </div>
      </div>
    </n-scrollbar>
  </div>
</template>

<style scoped>
#ResCheck {
  position: absolute;
  width: 100%;
  height: 100%;

  box-sizing: border-box;

  color: white;
}

#Content {
  width: 100%;
  height: 100%;
}

.errInfo {
  margin: 10px;
  border-radius: 1em;
  background-color: #3c3c3c;
  padding: 10px;
}

.errInfo > :nth-child(n) {
  display: flex;
}

.errInfo > :nth-child(n) > :first-child {
  flex: 0 0 100px;
}
</style>
