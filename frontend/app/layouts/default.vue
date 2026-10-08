<template>
  <div class="app-root">
    <!-- Window Bar (macOS Traffic Light Titlebar) -->
    <WindowBar />

    <div class="app-body">
      <!-- Sidebar Navigation with DB Status -->
      <Sidebar 
        :destPath="destPath"
        :currentTab="currentTab"
        @select-tab="handleSelectTab"
      />

      <!-- Content Area via Nuxt Slot -->
      <main class="main-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'

const route = useRoute()
const destPath = ref('~/Pictures/PhotoBackup')

const currentTab = computed(() => {
  if (route.path === '/gallery') return 'gallery'
  if (route.path === '/history') return 'history'
  if (route.path === '/settings') return 'settings'
  return 'backup'
})

function handleSelectTab(tab: string) {
  if (tab === 'backup') navigateTo('/')
  else if (tab === 'gallery') navigateTo('/gallery')
  else if (tab === 'history') navigateTo('/history')
  else if (tab === 'settings') navigateTo('/settings')
}

onMounted(async () => {
  if (typeof (window as any)?.go?.main?.App?.GetDestPath === 'function') {
    const dst = await (window as any).go.main.App.GetDestPath()
    if (dst) destPath.value = dst
  }
})
</script>

<style scoped>
.app-root {
  display: flex;
  flex-direction: column;
  width: 100vw;
  height: 100vh;
  overflow: hidden;
  background-color: #F4F5FA;
}

.app-body {
  display: flex;
  flex: 1;
  height: calc(100vh - 36px);
  overflow: hidden;
}

.main-content {
  flex: 1;
  height: 100%;
  overflow: hidden;
  position: relative;
}
</style>
