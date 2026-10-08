<template>
  <div class="app-root">
    <!-- Window Bar (macOS Traffic Light Titlebar) -->
    <WindowBar />

    <div class="app-body">
      <!-- Sidebar Navigation via NuxtLink -->
      <Sidebar :destPath="destPath" />

      <!-- Content Area rendered by Nuxt Routing -->
      <main class="main-content">
        <slot />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'

const destPath = ref('~/Pictures/PhotoBackup')

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
