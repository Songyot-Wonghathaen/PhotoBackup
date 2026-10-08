<template>
  <div class="settings-view">
    <!-- Notice Banner that this is a Mockup of a friend's part -->
    <div class="mock-notice">
      <div class="notice-badge">UI Mockup</div>
      <span class="notice-text">
        หน้านี้เป็นภาพตัวอย่าง (ส่วนของเพื่อน): การตั้งค่าระบบและการเชื่อมต่อ AI Model (Settings)
      </span>
    </div>

    <!-- Header -->
    <header class="view-header">
      <div class="header-text">
        <h1 class="view-title">ตั้งค่า</h1>
        <p class="view-subtitle">กำหนดค่าการทำงานของระบบสำรองข้อมูล, การตรวจจับไฟล์ซ้ำ และ AI</p>
      </div>
    </header>

    <div class="settings-container">
      <!-- Section 1: Backup Engine -->
      <div class="settings-card">
        <h3 class="card-title">
          <span class="card-icon">⚡</span>
          ระบบสำรองและคัดลอกไฟล์ (Part B)
        </h3>
        
        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">ความเร็วการทำงาน Concurrency Worker Pool</span>
            <span class="setting-desc">จำนวน Goroutines สำหรับคัดลอกไฟล์แบบขนานพร้อมกัน</span>
          </div>
          <div class="setting-control">
            <span class="badge-value">16 Workers</span>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">ขนาด Stream Buffer</span>
            <span class="setting-desc">ขนาดบัฟเฟอร์หน่วยความจำสำหรับการย้ายข้ามไดรฟ์ (CopyBuffer)</span>
          </div>
          <div class="setting-control">
            <span class="badge-value">128 KB</span>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">ตรวจสอบความปลอดภัยไฟล์ซ้ำ (Duplicate Guard)</span>
            <span class="setting-desc">ตรวจสอบไฟล์ซ้ำทั้งในดิสก์ปลายทางและฐานข้อมูลด้วย SHA256</span>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input type="checkbox" checked disabled>
              <span class="slider"></span>
            </label>
          </div>
        </div>
      </div>

      <!-- Section 2: Database -->
      <div class="settings-card">
        <h3 class="card-title">
          <span class="card-icon">🗄️</span>
          ฐานข้อมูล (SQLite + GORM)
        </h3>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">SQLite WAL Mode</span>
            <span class="setting-desc">Write-Ahead Logging เพิ่มความเร็วในการเขียนข้อมูลพร้อมกัน</span>
          </div>
          <div class="setting-control">
            <span class="status-tag active">เปิดใช้งาน (WAL)</span>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">Batch Insert Size</span>
            <span class="setting-desc">บันทึกข้อมูลรูปภาพลงฐานข้อมูลแบบแบทช์เพื่อประสิทธิภาพสูงสุด</span>
          </div>
          <div class="setting-control">
            <span class="badge-value">1,000 แถว / รอบ</span>
          </div>
        </div>
      </div>

      <!-- Section 3: AI Service (Friend's part) -->
      <div class="settings-card">
        <h3 class="card-title">
          <span class="card-icon">🤖</span>
          การวิเคราะห์ภาพด้วย AI (ส่วนของเพื่อน)
        </h3>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">วิเคราะห์ภาพอัตโนมัติหลังสำรอง</span>
            <span class="setting-desc">ส่งรูปภาพไปสร้างคำอธิบายภาพและ Tags ทันทีที่ย้ายเสร็จ</span>
          </div>
          <div class="setting-control">
            <label class="toggle-switch">
              <input type="checkbox" checked>
              <span class="slider"></span>
            </label>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <span class="setting-name">AI Model</span>
            <span class="setting-desc">โมเดลที่ใช้สร้างคำอธิบายภาพและ Tag</span>
          </div>
          <div class="setting-control">
            <select class="select-box" disabled>
              <option>Moondream2 / BLIP-2</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.settings-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 24px 28px 16px 28px;
  box-sizing: border-box;
  background-color: #F4F5FA;
  overflow-y: auto;
  user-select: none;
  font-family: 'Prompt', 'Inter', sans-serif;
}

.mock-notice {
  background: #EEF2FF;
  border: 1px dashed #6366F1;
  border-radius: 10px;
  padding: 8px 14px;
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
}

.notice-badge {
  background: #4F46E5;
  color: #FFFFFF;
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 6px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.notice-text {
  font-size: 12.5px;
  color: #3730A3;
  font-weight: 500;
}

.view-header {
  margin-bottom: 20px;
}

.view-title {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 4px 0;
}

.view-subtitle {
  font-size: 13px;
  color: #6B7280;
  margin: 0;
}

.settings-container {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-width: 860px;
}

.settings-card {
  background: #FFFFFF;
  border: 1px solid #E5E7EB;
  border-radius: 14px;
  padding: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.02);
}

.card-title {
  font-size: 15px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 14px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 0;
  border-bottom: 1px solid #F3F4F6;
}

.setting-row:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.setting-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-name {
  font-size: 13.5px;
  font-weight: 600;
  color: #1F2937;
}

.setting-desc {
  font-size: 12px;
  color: #6B7280;
}

.badge-value {
  background: #EEF2FF;
  color: #4F46E5;
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 6px;
  font-family: monospace;
}

.status-tag {
  font-size: 12px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 9999px;
}

.status-tag.active {
  background: #ECFDF5;
  color: #059669;
}

.select-box {
  background: #F9FAFB;
  border: 1px solid #D1D5DB;
  border-radius: 6px;
  padding: 6px 12px;
  font-size: 12.5px;
  font-family: inherit;
  color: #374151;
}

/* Switch */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #E5E7EB;
  transition: .3s;
  border-radius: 24px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .3s;
  border-radius: 50%;
  box-shadow: 0 1px 3px rgba(0,0,0,0.15);
}

input:checked + .slider {
  background-color: #4F46E5;
}

input:checked + .slider:before {
  transform: translateX(20px);
}
</style>
