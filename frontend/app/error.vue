<template>
  <div class="error-page">
    <div class="error-card">
      <div class="error-icon">⚠️</div>
      <h1 class="error-title">เกิดข้อผิดพลาด</h1>
      <p class="error-message">{{ error?.message || 'ไม่สามารถโหลดหน้านี้ได้' }}</p>
      <button class="btn-home" @click="handleClearError">กลับสู่หน้าหลัก</button>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps({
  error: Object
})

const handleClearError = () => {
  clearError({ redirect: '/' })
}
</script>

<style scoped>
.error-page {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background-color: #F4F5FA;
  font-family: 'Prompt', 'Inter', sans-serif;
}

.error-card {
  background: #FFFFFF;
  padding: 32px 40px;
  border-radius: 16px;
  box-shadow: 0 10px 25px rgba(0,0,0,0.05);
  text-align: center;
  max-width: 420px;
}

.error-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.error-title {
  font-size: 20px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 8px 0;
}

.error-message {
  font-size: 13.5px;
  color: #6B7280;
  margin: 0 0 20px 0;
}

.btn-home {
  background: #4F46E5;
  color: #FFFFFF;
  border: none;
  border-radius: 8px;
  padding: 10px 24px;
  font-size: 13.5px;
  font-weight: 600;
  cursor: pointer;
}
</style>
