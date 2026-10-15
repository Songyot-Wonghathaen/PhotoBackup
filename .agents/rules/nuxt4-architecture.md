# กฎระเบียบและแนวทางการพัฒนา PhotoVault (PhotoBackup)

## 1. กฎเหล็กการเปลี่ยนหน้าและการวางเส้นทาง (Strict Sub-Folder Routing Rule)
**การนำทางและสร้างเส้นทางทั้งหมด ต้องใช้ระบบ Sub-folder Route ของ Nuxt ตามที่เอกสารการเรียนการสอนระบุอย่างเคร่งครัด**:
- หน้าหลักเริ่มต้น:
  - `pages/index.vue` -> เส้นทาง `/` (หน้าหลักเริ่มต้น: สำรองรูปภาพ Part B)
- หน้าจอระบบอื่น ๆ ต้องสร้างเป็นโฟลเดอร์ย่อยและมีไฟล์ `index.vue` เสมอ (Sub-folder Route):
  - `pages/gallery/index.vue` -> เส้นทาง `/gallery` (หน้าแกลเลอรี & ค้นหา)
  - `pages/history/index.vue` -> เส้นทาง `/history` (หน้าประวัติ & ตรวจสอบ)
  - `pages/settings/index.vue` -> เส้นทาง `/settings` (หน้าตั้งค่า)
  - *(รองรับ Dynamic Parameter ในอนาคต เช่น `pages/gallery/[idx].vue` ด้วย `$route.params.idx`)*
- **การเปลี่ยนหน้าต้องใช้แท็ก `<NuxtLink to="...">`** สำหรับ Client-side Navigation เพื่อความลื่นไหลและไม่ต้องรีเฟรชหน้าต่างแอปพลิเคชัน
- ใน `app.vue` ต้องมีแท็ก `<NuxtPage />` (ครอบด้วย `<NuxtLayout>` ร่วมกับ `layouts/default.vue`)
- ชื่อโฟลเดอร์และไฟล์ภายใต้ `pages` ต้องเป็นตัวพิมพ์เล็ก (lowercase/kebab-case) ทั้งหมด

---

## 2. โครงสร้างโฟลเดอร์ Nuxt 4 (Strict Architecture)
```text
root/
├── app/
│   ├── assets/        # Processed static assets (CSS, fonts, images)
│   ├── components/    # Reusable Vue components (Auto-imported)
│   ├── composables/   # Custom Composition API hooks (useBackup.ts)
│   ├── layouts/       # Shared UI layouts (layouts/default.vue)
│   ├── middleware/    # Navigation middleware
│   ├── pages/         # Page routes แบบ Sub-folder Route
│   │   ├── index.vue            # Route: / (หน้าหลัก สำรองรูปภาพ)
│   │   ├── gallery/
│   │   │   └── index.vue        # Route: /gallery (แกลเลอรี & ค้นหา)
│   │   ├── history/
│   │   │   └── index.vue        # Route: /history (ประวัติ & ตรวจสอบ)
│   │   └── settings/
│   │       └── index.vue        # Route: /settings (ตั้งค่าระบบ)
│   ├── plugins/       # Vue plugins & runtime libraries
│   ├── utils/         # Helper functions (formatters)
│   ├── app.vue        # Root component (<NuxtLayout><NuxtPage /></NuxtLayout>)
│   ├── error.vue      # Custom error UI page
│   └── app.config.ts  # Reactive app-level configuration
├── server/            # Backend server routes
├── public/            # Static files (favicon.ico, robots.txt)
├── modules/           # Local Nuxt modules
├── nuxt.config.ts     # Config หลัก (ssr: false, nitro.output.publicDir: 'dist')
├── package.json       # Dependencies & Scripts
├── tsconfig.json      # TypeScript configuration
└── .nuxtignore        # Ignore patterns
```

---

## 3. แผนงานและรายการฟีเจอร์ทั้งหมด (Feature Roadmap & Ownership)

### A. โครงโปรเจกต์และฐานข้อมูล
- [x] ตั้งโปรเจกต์ Wails + Nuxt ให้รันได้บน Windows และ macOS
- [x] เชื่อม SQLite ผ่าน GORM และสร้างตาราง `photos`, `tags`, `photo_tags` (WAL mode เปิดใช้งานแล้ว)
- [ ] Model และฟังก์ชัน CRUD ของ 3 ตาราง (Model ของ SQLite ทำแล้ว เหลือฟังก์ชัน CRUD)

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

### C. AI วิเคราะห์ภาพ (งานของ ป้อ: ฟีเจอร์ข้อ 12 - 14)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 12. ส่งรูปให้ AI ได้คำอธิบายและ Tag ตอนสำรอง (วิเคราะห์ครั้งเดียว)
- [ ] 13. บันทึกคำอธิบายและ Tag ลง DB
- [ ] 14. จัดการกรณี AI ล้มเหลว (ย้ายไฟล์ตามปกติ ปล่อยคำอธิบายว่าง)

### D. ค้นหาและแสดงภาพ (งานของ Pun: ฟีเจอร์ข้อ 15 - 19)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 15. ค้นหาจากคำอธิบาย (ช่องกรอกข้อความ พิมพ์คำค้นหาได้ตามต้องการ)
- [ ] 16. Tag Cloud (ขนาดตัวอักษรตามจำนวนรูป)
- [ ] 17. ค้นหาด้วยการคลิก Tag และแสดงรูปหลายภาพ
- [ ] 18. หน้ารายละเอียดภาพ: ภาพใหญ่, คำอธิบาย, Tag, ข้อมูลไฟล์
- [ ] 19. แก้ไขคำอธิบาย และเพิ่ม/ลบ Tag เอง

### E. ประวัติและตรวจสอบ (งานของ Per: ฟีเจอร์ข้อ 20 - 25)
*สถานะ: ยังไม่มีระบบ ให้ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 20. ตารางประวัติจาก DB เฉพาะปลายทางปัจจุบัน
- [ ] 21. ตรวจความถูกต้อง (`/check`): เทียบ DB กับไฟล์จริง รายงานไฟล์สูญหาย
- [ ] 22. สถิติ (ทั้งหมด / ผ่าน / สูญหาย / ลบแล้ว)
- [ ] 23. ลบไฟล์ในปลายทางและอัปเดตสถานะใน DB
- [ ] 24. ดูประวัติของปลายทางอื่นได้ (`/list db <path>`) ผ่านปุ่ม "เปลี่ยนปลายทาง" ในหน้า 5
- [ ] 25. ลบไฟล์ทั้งหมดในปลายทาง (`/delete all`) ไม่ใช่แค่ทีละไฟล์

---

## 4. กฎความร่วมมือในทีมและการควบคุมเวอร์ชัน (Git Rules)
- Part B พัฒนาระบบจริง ส่วน Part C, D, E ทำเป็น UI Mockup รอเพื่อน
- ห้ามดัดแปลงหรือลบ 3 ตารางในฐานข้อมูล
- ทำงานและ Push การเปลี่ยนแปลงขึ้น Branch **`featureB`** เท่านั้น
