<template>
  <div class="agent-redirect">正在打开 XyAgent…</div>
</template>

<script setup>
import { onMounted } from 'vue'

defineOptions({ name: 'XyAgent' })
onMounted(() => window.location.assign('/agent/'))
</script>

<style scoped>
.agent-redirect {
  min-height: 50vh;
  display: grid;
  place-items: center;
  color: var(--el-text-color-secondary);
}
</style>
