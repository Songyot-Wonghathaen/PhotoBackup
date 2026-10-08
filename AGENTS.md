# กฎระเบียบ ข้อกำหนด และโครงสร้างสถาปัตยกรรม PhotoVault

เอกสารนี้เป็นข้อกำหนดและกฎอย่างเคร่งครัดสำหรับการพัฒนาแอปพลิเคชัน PhotoVault (Wails + Nuxt 4 + SQLite GORM)

---

## 1. แผนงานและรายการฟีเจอร์ทั้งหมด (Feature Roadmap & Ownership)

### A. โครงโปรเจกต์และฐานข้อมูล
- [x] ตั้งโปรเจกต์ Wails + Nuxt ให้รันได้บน Windows และ macOS
- [x] เชื่อม SQLite ผ่าน GORM และสร้างตาราง `photos`, `tags`, `photo_tags` (WAL mode เปิดใช้งานแล้ว)
- [ ] Model และฟังก์ชัน CRUD ของ 3 ตาราง (Model ของ SQLite ทำแล้ว เหลือฟังก์ชัน CRUD เพิ่มเติม)

---

### B. สำรองรูปภาพ (งานของ Keen: ฟีเจอร์ข้อ 4 - 11) [รับผิดชอบโดย Keen]
*ใช้เทคนิคเทียบเท่าข้อสอบกลางภาค (เต็ม 100%)*
- [x] **4. เลือกโฟลเดอร์ต้นทาง** (Flash Drive / โทรศัพท์) และแสดงรายการรูปภาพ
- [x] **5. เลือกโฟลเดอร์ปลายทาง** พร้อมสแกนตรวจไฟล์สูญหายทันที (`CheckIntegrity`)
- [x] **6. ย้ายไฟล์** (เลือกบางไฟล์ / ทั้งหมด): ลอง `os.Rename` ก่อน หากข้ามไดรฟ์ให้ Stream `io.CopyBuffer` 128KB แล้วลบต้นทาง
- [x] **7. กันไฟล์ซ้ำ**: ตรวจทั้งในโฟลเดอร์ปลายทางและในฐานข้อมูล (ชื่อไฟล์ + SHA256) พร้อมแจ้งเตือน
- [x] **8. จับเวลาการสำรองทั้งหมด** และแสดงผล (ความละเอียดระดับ Millisecond)
- [x] **9. ความเร็วสูง (High Performance)**:
  - ประมวลผลหลายไฟล์พร้อมกันด้วย Concurrency Worker Pool (16 goroutines)
  - บันทึกลงฐานข้อมูลเป็นแบทช์ด้วย GORM `CreateInBatches(1000)`
  - SQLite WAL mode (`PRAGMA journal_mode = WAL`, `cache_size = -64000`)
- [x] **10. แสดงไฟล์จริงในโฟลเดอร์ปลายทาง** (`/list dest`)
- [x] **11. State Guard**: ป้องกันกรณีที่ยังไม่ได้ตั้งปลายทางหรือต้นทาง ให้แจ้งเตือนก่อน ป้องกันโปรแกรม Crash

---

### C. AI วิเคราะห์ภาพ (งานของ ป้อ: ฟีเจอร์ข้อ 12 - 14)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 12. ส่งรูปให้ AI ได้คำอธิบายและ Tag ตอนสำรอง (วิเคราะห์ครั้งเดียว)
- [ ] 13. บันทึกคำอธิบายและ Tag ลง DB
- [ ] 14. จัดการกรณี AI ล้มเหลว (ย้ายไฟล์ตามปกติ ปล่อยคำอธิบายว่าง)

---

### D. ค้นหาและแสดงภาพ (งานของ Pun: ฟีเจอร์ข้อ 15 - 19)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 15. ค้นหาจากคำอธิบาย (ช่องกรอกข้อความ พิมพ์คำค้นหาได้ตามต้องการ)
- [ ] 16. Tag Cloud (ขนาดตัวอักษรตามจำนวนรูป)
- [ ] 17. ค้นหาด้วยการคลิก Tag และแสดงรูปหลายภาพ
- [ ] 18. หน้ารายละเอียดภาพ: ภาพใหญ่, คำอธิบาย, Tag, ข้อมูลไฟล์
- [ ] 19. แก้ไขคำอธิบาย และเพิ่ม/ลบ Tag เอง

---

### E. ประวัติและตรวจสอบ (งานของ Per: ฟีเจอร์ข้อ 20 - 25)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 20. ตารางประวัติจาก DB เฉพาะปลายทางปัจจุบัน
- [ ] 21. ตรวจความถูกต้อง (`/check`): เทียบ DB กับไฟล์จริง รายงานไฟล์สูญหาย
- [ ] 22. สถิติ (ทั้งหมด / ผ่าน / สูญหาย / ลบแล้ว)
- [ ] 23. ลบไฟล์ในปลายทางและอัปเดตสถานะใน DB
- [ ] 24. ดูประวัติของปลายทางอื่นได้ (`/list db <path>`) ผ่านปุ่ม "เปลี่ยนปลายทาง" ในหน้า 5
- [ ] 25. ลบไฟล์ทั้งหมดในปลายทาง (`/delete all`) ไม่ใช่แค่ทีละไฟล์

---

## 2. กฎการวางโครงสร้างโฟลเดอร์ (Nuxt 4 Strict Architecture)

โครงสร้างโฟลเดอร์ฝั่ง Frontend ต้องจัดตามนี้อย่างเคร่งครัด:

```text
root/
├── app/
│   ├── assets/        # Processed static assets (CSS, fonts, images)
│   ├── components/    # Reusable Vue components (Auto-imported)
│   ├── composables/   # Custom Composition API hooks (useBackup.ts)
│   ├── layouts/       # Shared UI layouts (default.vue)
│   ├── middleware/    # Navigation middleware
│   ├── pages/         # Page routes (index.vue, gallery.vue, history.vue, settings.vue)
│   ├── plugins/       # Vue plugins & runtime libraries
│   ├── utils/         # Helper functions (formatters, calculations)
│   ├── app.vue        # Root component (<NuxtLayout><NuxtPage /></NuxtLayout>)
│   ├── error.vue      # Custom error UI page
│   └── app.config.ts  # Reactive app-level configuration
├── server/            # Backend server routes (ถ้ามี)
├── public/            # Public static files (favicon.ico, robots.txt)
├── modules/           # Local Nuxt modules
├── nuxt.config.ts     # Config หลัก (ssr: false, nitro.output.publicDir: 'dist')
├── package.json       # Dependencies & Scripts
├── tsconfig.json      # TypeScript configuration
├── .nuxtignore        # Files excluded during build
└── node_modules/
```

---

## 3. กฎความร่วมมือในทีมและการควบคุมเวอร์ชัน (Git & Team Rules)

1. **การจำกัดขอบเขต (Strict Scope Rule)**:
   - ส่วนที่เรารับผิดชอบจริงคือ **Part B (ฟีเจอร์ข้อ 4 - 11)** เท่านั้น
   - ส่วนที่ยังไม่มีระบบ (Part C, D, E) ของเพื่อน **ให้ทำเป็นหน้าภาพตัวอย่าง (UI Mockup) ไปก่อน** เพื่อให้เห็นการทำงานภาพรวมโดยไม่ไปทับซ้อนกับระบบของเพื่อน
2. **ห้ามแก้ไข Schema ฐานข้อมูล 3 ตารางหลัก**:
   - ตารางที่เพื่อนออกแบบมี 3 ตารางคือ `photos`, `tags`, `photo_tags` ห้ามสร้างตารางอื่นที่ไม่เกี่ยวข้องหรือลบฟิลด์เดิม
3. **การส่งโค้ดผ่าน Git**:
   - ห้าม Commit หรือ Push ขึ้น Branch `main` หรือ `develop` โดยตรงเด็ดขาด
   - งานทั้งหมดต้องทำและ Push ขึ้นที่ Branch **`featureB`** เท่านั้น
