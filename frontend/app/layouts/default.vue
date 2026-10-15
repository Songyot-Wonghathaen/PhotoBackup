<template>
  <div class="app-root">
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
import { onMounted } from 'vue'
import { useBackup } from '~/composables/useBackup'

const { destPath, initBackupState } = useBackup()

onMounted(() => {
  initBackupState()
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
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

.main-content {
  flex: 1;
  height: 100%;
  overflow: hidden;
  position: relative;
}
</style>
