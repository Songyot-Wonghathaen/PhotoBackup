<template>
  <div class="window-bar" style="--wails-draggable:drag">
    <div class="traffic-lights">
      <span class="dot red"></span>
      <span class="dot yellow"></span>
      <span class="dot green"></span>
    </div>
    <div class="window-title">PhotoVault — Wails + Nuxt</div>
    <div class="window-actions"></div>
  </div>
</template>

<style scoped>
.window-bar {
  height: 36px;
  background-color: #E5E7EB;
  border-bottom: 1px solid #D1D5DB;
  display: flex;
  align-items: center;
  padding: 0 16px;
  user-select: none;
  position: relative;
  flex-shrink: 0;
  z-index: 100;
}

.traffic-lights {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.dot.red {
  background-color: #EF4444;
}

.dot.yellow {
  background-color: #F59E0B;
}

.dot.green {
  background-color: #22C55E;
}

.window-title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-family: 'Prompt', 'Inter', sans-serif;
  font-size: 13px;
  font-weight: 500;
  color: #374151;
  pointer-events: none;
  letter-spacing: -0.1px;
}

.window-actions {
  margin-left: auto;
}
</style>
