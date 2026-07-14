<script setup lang="ts">
const props = defineProps<{
  checked: boolean
  label: string
  tooltip?: string
}>()

const emit = defineEmits<{
  (e: 'update:checked', value: boolean): void
}>()
</script>

<template>
  <n-tooltip v-if="tooltip" trigger="hover">
    <template #trigger>
      <n-checkbox
        :checked="checked"
        @update:checked="emit('update:checked', $event)"
      >
        {{ label }}
      </n-checkbox>
    </template>
    {{ tooltip }}
  </n-tooltip>
  <n-checkbox
    v-else
    :checked="checked"
    @update:checked="emit('update:checked', $event)"
  >
    {{ label }}
  </n-checkbox>
</template>
