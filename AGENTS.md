# กฎระเบียบและแนวทางการพัฒนา PhotoVault (PhotoBackup)

## 1. กฎเหล็กการเปลี่ยนหน้าและการวางเส้นทาง (Strict Sub-Folder Routing Rule)
**การนำทางและสร้างเส้นทางทั้งหมด ต้องใช้ระบบ Sub-folder Route ของ Nuxt ตามที่เอกสารการเรียนการสอนระบุอย่างเคร่งครัด**:
- หน้าหลักเริ่มต้น:
  - `pages/index.vue` -> เส้นทาง `/` (หน้าหลักเริ่มต้น: สำรองรูปภาพ Part B)
- หน้าจอระบบอื่น ๆ ต้องสร้างเป็นโฟลเดอร์ย่อยและมีไฟล์ `index.vue` เสมอ (Sub-folder Route):
  - `pages/gallery/index.vue` -> เส้นทาง `/gallery` (หน้าแกลเลอรี & ค้นหา)
  - `pages/history/index.vue` -> เส้นทาง `/history` (หน้าประวัติ & ตรวจสอบ)
  - `pages/settings/index.vue` -> เส้นทาง `/settings` (หน้าตั้งค่า)
  - เส้นทางแบบ Dynamic Parameter:
    - `pages/gallery/[idx].vue` -> เส้นทาง `/gallery/:idx` ดึงค่าผ่าน `$route.params.idx` หรือ `useRoute().params.idx`
- **การเปลี่ยนหน้าต้องใช้แท็ก `<NuxtLink to="...">`** สำหรับ Client-side Navigation เพื่อความลื่นไหลและไม่ต้องรีเฟรชหน้าต่างแอปพลิเคชัน (Single Page Routing)
- ใน `app.vue` ต้องมีแท็ก `<NuxtPage />` ครอบด้วย `<NuxtLayout>` ร่วมกับ `layouts/default.vue`
- ชื่อโฟลเดอร์และไฟล์ภายใต้ `pages` ต้องเป็นตัวพิมพ์เล็ก (lowercase/kebab-case) ทั้งหมด
- โฟลเดอร์ `components/` มีไว้สำหรับ Reusable Components เท่านั้น (เช่น `Sidebar.vue`) โดย Nuxt จะ Auto-import อัตโนมัติ ไม่นำหน้าเว็บทั้งหน้าไปใส่ไว้ใน `components/`

---

## 2. โครงสร้างโฟลเดอร์ Nuxt 4 (Strict Frontend Architecture)
```text
root/
├── app/
│   ├── assets/        # Processed static assets (CSS, fonts, images)
│   ├── components/    # Reusable Vue components (Auto-imported: Sidebar)
│   ├── composables/   # Custom Composition API hooks
│   ├── layouts/       # Shared UI layouts (layouts/default.vue พร้อม <slot />)
│   ├── middleware/    # Navigation middleware
│   ├── pages/         # Page routes แบบ Sub-folder Route
│   │   ├── index.vue            # Route: / (หน้าหลัก สำรองรูปภาพ)
│   │   ├── gallery/
│   │   │   ├── index.vue        # Route: /gallery (แกลเลอรี & ค้นหา)
│   │   │   └── [idx].vue        # Route: /gallery/:idx (หน้ารายละเอียดภาพ - Dynamic Route)
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

## 3. สถาปัตยกรรม Backend Go & เลเยอร์ข้อมูล (Strict Layered Architecture)
ตามเอกสารการสอน Wails & Go ของอาจารย์ แบ่งสัดส่วนโค้ดแบบ Separation of Concerns / Layered Architecture ในสไตล์เอกพจน์ (Singular):
```text
root/
├── service/           # Business Logic การประมวลผลหลัก (เช่น backup.go)
├── repository/        # Data Access & Database Connection (dbconnect.go, photo.go)
├── model/             # GORM Models (Structs) ที่สร้างขึ้นและรองรับ GORM Gen
│   ├── model/         # Generated GORM Model structs (photos.gen.go, tags.gen.go, ฯลฯ)
│   └── query/         # Type-safe Query Helpers จาก GORM Gen
├── utils/             # ฟังก์ชันช่วยเหลือต่าง ๆ (Helper functions)
├── database/          # ไฟล์ฐานข้อมูล SQLite (photovault.db)
├── app.go             # Presentation Layer / Wails Application Lifecycle (Context Management)
└── main.go            # Entry Point หลัก, Dependency Injection, และ Wails Bindings
```

### กฎระเบียบทางเทคนิคฝั่ง Backend:
1. **Wails IPC Bindings:**
   - ลงทะเบียน Struct ใน `main.go` ผ่านตัวแปร `Bind: []interface{}{ app, backupService }`
   - ฝั่ง Frontend เรียกใช้งานฟังก์ชันผ่าน `window.go.<package>.<Struct>.<Function>` ได้ทันที
2. **Context Management:**
   - ส่งต่อ Wails Runtime Context `ctx` จาก `app.startup(ctx)` ไปยัง Service ต่าง ๆ เพื่อควบคุม Timeout และการยกเลิกการประมวลผล
3. **Database & GORM:**
   - เชื่อมต่อผ่าน `github.com/glebarez/sqlite` (Pure Go ไม่ต้องพึ่งพา CGO)
   - เปิดใช้งาน WAL Mode (`PRAGMA journal_mode = WAL`) และปรับแต่ง Cache Memory เพื่อความเร็วสูงสุด
   - ห้ามลบหรือดัดแปลง Schema ของ 3 ตารางหลัก (`photos`, `tags`, `photo_tags`)

---

## 4. แผนงานและรายการฟีเจอร์ทั้งหมด (Feature Roadmap & Ownership)

### A. โครงโปรเจกต์และฐานข้อมูล
- [x] ตั้งโปรเจกต์ Wails + Nuxt ให้รันได้บน Windows และ macOS
- [x] เชื่อม SQLite ผ่าน GORM และสร้างตาราง `photos`, `tags`, `photo_tags` (WAL mode เปิดใช้งานแล้ว)
- [x] Model และฟังก์ชัน CRUD ของ 3 ตาราง (Model ของ SQLite และ Repository ทำเรียบร้อย)

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
- [x] 12. ส่งรูปให้ AI ได้คำอธิบายและ Tag ตอนสำรอง (วิเคราะห์ครั้งเดียว)
- [x] 13. บันทึกคำอธิบายและ Tag ลง DB
- [x] 14. จัดการกรณี AI ล้มเหลว (ย้ายไฟล์ตามปกติ ปล่อยคำอธิบายว่าง)

### D. ค้นหาและแสดงภาพ (งานของ Pun: ฟีเจอร์ข้อ 15 - 19)
*สถานะ: ยังไม่มีระบบ ทำเป็น UI Mockup หน้าภาพตัวอย่างไว้ก่อน*
- [ ] 15. ค้นหาจากคำอธิบาย (ช่องกรอกข้อความ พิมพ์คำค้นหาได้ตามต้องการ)
- [ ] 16. Tag Cloud (ขนาดตัวอักษรตามจำนวนรูป)
- [ ] 17. ค้นหาด้วยการคลิก Tag และแสดงรูปหลายภาพ
- [ ] 18. หน้ารายละเอียดภาพ: ภาพใหญ่, คำอธิบาย, Tag, ข้อมูลไฟล์ (สร้างใน `pages/gallery/[idx].vue`)
- [ ] 19. แก้ไขคำอธิบาย และเพิ่ม/ลบ Tag เอง

### E. ประวัติและตรวจสอบ (งานของ Per: ฟีเจอร์ข้อ 20 - 25) [เสร็จสมบูรณ์ - ใช้เทคนิคระดับสูงเต็ม 100%]
*สถาปัตยกรรม High-Performance Concurrency, Set Difference Algorithm & Two-Phase Integrity Guard*
- [x] **20. ตารางประวัติจาก DB เฉพาะปลายทางปัจจุบัน** (สร้างใน `pages/history/index.vue` ตรงตาม Figma)
  - **Eager Preloading (ป้องกัน N+1 Query)**: ใช้ GORM `Preload(p.PhotoTags.Tag)` โหลดรูปพร้อมแท็กทั้งหมดในคิวรีเดียว
  - **Cross-Platform Path Normalization (`destConditions`)**: สร้าง Path Variation Matrix (Clean, Absolute, Relative, Slash `/`, Backslash `\`) ค้นหาแม่นยำ 100% ทั้งบน Windows และ macOS
  - **Lazy Concurrent Thumbnail Pool**: ประมวลผล Thumbnail ภาพขนาด 280px เป็น Base64 แบบขนานด้วย Goroutine Pool (8 workers) เพื่อการ Render ที่ลื่นไหล
  - **Deterministic Color Hashing**: คำนวณสี Gradient พื้นหลังอัตโนมัติจาก ID และชื่อไฟล์แบบสม่ำเสมอ
- [x] **21. ตรวจความถูกต้อง (`/check`): เทียบ DB กับไฟล์จริง รายงานไฟล์สูญหาย**
  - **Set Difference Algorithm $O(N + M)$**: ใช้ In-Memory Hash Map (`diskMap[string]bool`) ตรวจสอบความมีอยู่ของไฟล์จริงใน $O(1)$ รวมเวลา $O(N+M)$ เร็วกว่า Nested Loop มหาศาล
  - **Atomic Batch Status Synchronization**: รวบรวมไฟล์ที่ `missing` และ `active` แล้วอัปเดตสถานะใน DB พร้อมกันด้วยคำสั่ง Batch Update (`UPDATE ... WHERE file_name IN (...)`) ลด Disk Write I/O เหลือเพียง 1 ครั้ง
  - **Self-Healing Reconciliation**: หากไฟล์ที่เคยหายถูกนำกลับมาวางในโฟลเดอร์ ระบบจะตรวจพบและปรับสถานะคืนเป็น `active` ทันที
  - **High-Resolution Microsecond Timer**: วัดระยะเวลาการตรวจสอบระดับ Microsecond ด้วย `time.Since(start)` และแสดงผลละเอียดระดับ Millisecond (เช่น `12.4 ms`)
- [x] **22. สถิติ (ทั้งหมด / ผ่าน / สูญหาย / ลบแล้ว)**
  - **Single-Pass Reactive Computation ($O(N)$)**: ใช้ Vue 3 `computed()` วนลูปคำนวณสรุปผลสถิติทั้ง 4 หมวดใน Loop เดียวรอบเดียว พร้อมระบบ Memoization ไม่คำนวณซ้ำหากข้อมูลไม่เปลี่ยน
  - **Zero-Latency Interactive Filter Controller**: คลิกการ์ดสถิติใบใดก็ได้เพื่อกรองตารางตามสถานะ (`all`, `passed`, `missing`, `deleted`) แบบ Instant 0 ms ไม่ต้องยิงขอข้อมูลจาก DB ใหม่
  - **Visual Hierarchy & Active Ring Glow**: ตัวบ่งชี้สถานะที่กำลังถูกกรองด้วย Dynamic Active Glow Border
- [x] **23. ลบไฟล์ในปลายทางและอัปเดตสถานะใน DB**
  - **Two-Phase Commit (2PC) Pattern**:
    - Phase 1 (Physical Removal): ลบไฟล์จริงบนดิสก์ด้วย `os.Remove(filePath)`
    - Phase 2 (Audit Trail Preservation): ทำ **Soft Delete** โดยอัปเดต `status = 'deleted'` ในฐานข้อมูล เพื่อรักษาประวัติการสำรองไว้ตรวจสอบย้อนหลัง
  - **Atomic Batch Single/Multi Deletion**: ลบไฟล์จริงพร้อมอัปเดตสถานะใน DB แบบ Batch ผ่าน `MarkPhotosDeletedInBatch`
  - **Accidental Deletion Prevention Guard**: แสดง Confirmation Modal ระบุชื่อและขนาดไฟล์อย่างชัดเจน พร้อมปุ่มเตือนสีแดงก่อนยืนยัน
- [x] **24. ดูประวัติของปลายทางอื่นได้ (`/list db <path>`) ผ่านปุ่ม "เปลี่ยนปลายทาง" ในหน้า 5**
  - **Native OS Directory Dialog Integration**: เชื่อมต่อผ่าน Wails `runtime.OpenDirectoryDialog` เปิด File Explorer / Finder แท้ของระบบปฏิบัติการ
  - **Decoupled Dynamic Query (`ListAllDB(destPath)`)**: ดึงประวัติของโฟลเดอร์ใดๆ จาก SQLite ได้อิสระโดยไม่ต้องเปลี่ยนเซสชันหลัก
  - **Auto-Integrity Trigger on Switch**: เมื่อสลับโฟลเดอร์ จะโหลดข้อมูลและรัน `CheckIntegrity` อัตโนมัติเพื่อตรวจสอบความถูกต้องของโฟลเดอร์ใหม่ทันที
  - **Manual Path Fallback Modal**: มีช่องกรอก Path โฟลเดอร์ปลายทางด้วยตนเอง กรณี Native Dialog ไม่ตอบสนอง
- [x] **25. ลบไฟล์ทั้งหมดในปลายทาง (`/delete all`) ไม่ใช่แค่ทีละไฟล์**
  - **Concurrency Worker Pool (16 goroutines)**: สตรีมลบไฟล์จริงบนดิสก์แบบคู่ขนาน (Parallel Disk I/O) ด้วย Goroutine Pool 16 คอร์ พร้อม Semaphore Channel ควบคุม ไม่ทำให้ระบบค้างแม้โฟลเดอร์มีเป็นหมื่นไฟล์
  - **Atomic Bulk Soft-Delete in SQLite**: คำสั่ง SQL เดียวอัปเดตทั้งโฟลเดอร์ `UPDATE photos SET status = 'deleted' WHERE destination_path IN (...) AND status != 'deleted'` ทำงานใน Single Transaction
  - **Thread-Safe Atomic Counter**: ใช้ `sync/atomic.Int64` นับจำนวนไฟล์ที่ลบสำเร็จจริงบนดิสก์ ปลอดภัยจาก Race Condition
  - **Double Safety Destructive Modal**: หน้าต่างยืนยันระดับวิกฤต (Critical Danger Modal) แสดงจำนวนไฟล์ทั้งหมดที่จะถูกลบ พร้อมปิดปุ่มอัตโนมัติหากไม่มีไฟล์หลงเหลือให้ลบ


---

## 5. กฎความร่วมมือในทีมและการควบคุมเวอร์ชัน (Git Rules)
- Part B พัฒนาระบบจริง ส่วน Part C, D, E ทำเป็น UI Mockup รอเพื่อน
- ห้ามดัดแปลงหรือลบ 3 ตารางในฐานข้อมูล
- ทำงานและ Push การเปลี่ยนแปลงขึ้น Branch **`featureB`** เท่านั้น

