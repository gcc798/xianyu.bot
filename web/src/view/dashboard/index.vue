<template>
  <div class="h-full overflow-auto bg-slate-50/60 p-4 dark:bg-slate-900 lg:p-6">
    <section class="rounded-xl border border-slate-200/80 bg-white p-6 shadow-sm dark:border-slate-700 dark:bg-slate-800">
      <p class="text-xs tracking-[0.2em] text-slate-500 dark:text-slate-400">DASHBOARD</p>
      <h1 class="mt-2 text-2xl font-semibold text-slate-900 dark:text-slate-100">欢迎回来</h1>
      <p class="mt-2 text-sm text-slate-600 dark:text-slate-300">{{ today }} · 请选择要管理的功能</p>
    </section>

    <gva-card title="快捷入口" custom-class="mt-4 max-w-2xl">
      <gva-quick-link />
    </gva-card>
  </div>
</template>

<script setup>
  import { computed } from 'vue'
  import { GvaCard, GvaQuickLink } from './components'

  const today = computed(() => new Date().toLocaleDateString('zh-CN'))

  defineOptions({
    name: 'Dashboard'
  })
</script>
