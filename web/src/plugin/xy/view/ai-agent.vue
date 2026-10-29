<template>
  <div class="agent-shell">
    <section class="agent-empty">
      <div class="agent-mark">XY</div>
      <p class="eyebrow">XY AGENT</p>
      <h1>从一个明确目标开始</h1>
      <p>模型和能力配置就绪后，可在这里发起对话并调用已连接的能力。</p>
      <el-alert title="Agent Loop 尚未启用；当前阶段不会发送消息或调用工具。" type="info" :closable="false" show-icon />
      <el-button type="primary" plain @click="$router.push({ name: 'XyAISettings' })">前往基础配置</el-button>
    </section>
  </div>
</template>

<script setup>
defineOptions({ name: 'XyAgent' })
</script>

<style scoped>
.agent-shell { min-height: calc(100vh - 132px); display: grid; place-items: center; padding: 24px; border-radius: 10px; background: linear-gradient(145deg, var(--el-bg-color) 0%, var(--el-fill-color-light) 100%); }
.agent-empty { width: min(560px, 100%); text-align: center; }
.agent-mark { width: 62px; height: 62px; display: grid; place-items: center; margin: 0 auto 18px; border: 1px solid #0f8f86; border-radius: 16px; color: #0f8f86; font: 700 18px/1 ui-monospace, monospace; }
.eyebrow { margin: 0; color: #0f8f86; font-size: 11px; letter-spacing: .22em; }
h1 { margin: 8px 0 10px; font-size: clamp(26px, 4vw, 38px); color: var(--el-text-color-primary); }
.agent-empty > p:not(.eyebrow) { margin: 0 auto 22px; color: var(--el-text-color-secondary); }
.el-alert { margin-bottom: 18px; text-align: left; }
</style>
