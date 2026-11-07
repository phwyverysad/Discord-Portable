# Discord PTB Portable & WebSetup Suite (Multi-Profile + Vencord Edition)

> Discord PTB (Public Test Build) Portable with **Vencord Client Mod**, **Multi-Profile Data Isolation**, **Token Security Guard**, and a **Modern Lightweight WebSetup Suite**.

[![GitHub Release](https://img.shields.io/github/v/release/phwyverysad/Discord-Portable?style=flat-square)](https://github.com/phwyverysad/Discord-Portable/releases)
[![Build Status](https://github.com/phwyverysad/Discord-Portable/actions/workflows/build.yml/badge.svg)](https://github.com/phwyverysad/Discord-Portable/actions)

---

## ✨ Features (จุดเด่นและความสามารถ)

### 🎨 1. Modern WebSetup Installer Suite (ตัวติดตั้งออนไลน์ขนาดเล็ก)
- **Compact Binary:** คอมไพล์เป็นไฟล์ `.exe` ขนาดเล็ก (ประมาณ 7-8 MB) บรรจุหน้าต่างติดตั้งที่เขียนด้วย HTML/CSS/JS
- **Light White-Blue Curved Theme:** ดีไซน์สว่าง สะอาดตา ขาว-ฟ้าอ่อน ขอบโค้งมนนุ่มนวล (`border-radius: 20px`)
- **Zero Emojis (100% Inline SVG):** ปราศจากอิโมจิ ใช้กราฟิกเวกเตอร์ SVG คุณภาพสูง
- **Bilingual Support (TH / EN):** สลับภาษาไทย-อังกฤษได้ทันทีแบบเรียลไทม์ที่มุมบนขวา

### 🌈 2. Multi-Profile Concurrency (เปิดได้หลายจอพร้อมกัน แยกบัญชีเด็ดขาด)
- **12 Distinct Colored Profiles:** แสดงการ์ด 12 สี พร้อมไอคอน Discord จริงจาก `discord_icons` และตั้งชื่อตามสี
  - `01` - Discord Yellow (สีเหลือง - โปรไฟล์หลัก)
  - `02` - Discord Yellow Orange (สีส้มอมเหลือง)
  - `03` - Discord Orange (สีส้ม)
  - `04` - Discord Red Orange (สีส้มแดง)
  - `05` - Discord Red (สีแดง)
  - `06` - Discord Red Violet (สีม่วงอมแดง)
  - `07` - Discord Violet (สีม่วง)
  - `08` - Discord Blue Violet (สีม่วงอมน้ำเงิน)
  - `09` - Discord Blue (สีน้ำเงิน)
  - `10` - Discord Blue Green (สีเขียวอมน้ำเงิน)
  - `11` - Discord Green (สีเขียว)
  - `12` - Discord Yellow Green (สีเขียวตองอ่อน)
- **Single Install Location:** ติดตั้งตัวโปรแกรมหลักไว้ตำแหน่งเดียว (เช่น `%LOCALAPPDATA%\Programs\Discord-Portable`) ไม่เปลืองพื้นที่
- **100% Session & Data Isolation:** โฟลเดอร์ข้อมูล (`data/profiles/profile_<id>`) แยกขาดจากกัน เปิดพร้อมกันได้หลายหน้าจอโดยไม่แชร์ประวัติหรือโทเค็น
- **Desktop Shortcuts with Color Icons:** สร้างทางลัดบน Desktop แยกตามโปรไฟล์พร้อมไอคอนสีประจำตัว

### 🛡️ 3. Token & Credential Security Shield (ระบบป้องกันการดักจับข้อมูล)
- **Argument Sanitizer:** บล็อกและตัดคำสั่งสอดแนม `--remote-debugging-port`, `--inspect`, `--inspect-brk` ป้องกันการดักจับหน่วยความจำและ Token
- **Stealer Path Deflector:** ซ่อนและกักเก็บไฟล์เซสชัน LevelDB ไว้เฉพาะในไดเรกทอรี Portable ป้องกันมัลแวร์ทั่วไปที่กวาดหาจาก `%APPDATA%`
- **Core Integrity Check:** ตรวจสอบ SHA-256 Hash ของไฟล์หลัก ป้องกันการแอบฉีดโค้ดแปลกปลอม

### 🚀 4. Pro Features (ฟีเจอร์ระดับโปร)
- **Eco / Low-RAM Mode:** โหมดประหยัดแรม หน่วงเฟรมเรตเบื้องหลัง ช่วยลดการใช้แรมลง 40–60% เมื่อเปิดหลายจอ
- **Smart RPC Router:** กำหนด Profile 1 เป็น Primary Rich Presence (`discord-ipc-0`) สำหรับเกมและ Spotify ส่วนโปรไฟล์อื่นๆ แยกช่องทางป้องกันการชนกัน
- **Colored System Tray Icons:** ถาดมุมขวาล่างแสดงไอคอนสีตามโปรไฟล์ที่เปิดอยู่
- **Dual Install Mode:** เลือกระหว่าง Standard PC Install หรือ True USB Portable สำหรับเสียบแฟลชไดรฟ์
- **One-Click Backup & Restore:** สำรองและกู้คืนโปรไฟล์เป็นไฟล์ `.zip` ปลอดภัยในคลิกเดียว
- **Self-Updating Launcher:** ตรวจจับและอัปเดตไฟล์ `.exe` ตัวเองอัตโนมัติจาก GitHub Releases

---

## 🛠️ Build & Verification (การบิลด์และทดสอบ)

### รันการทดสอบ 5 ด่านก่อนส่งมอบ (Acceptance Gates):
```powershell
powershell -ExecutionPolicy Bypass -File scripts\verify_delivery.ps1
```

### คอมไพล์ด้วย Go:
```powershell
# คอมไพล์ตัว Launcher
go build -ldflags="-s -w" -o DiscordPTB-Portable.exe main.go

# คอมไพล์ตัว WebSetup Suite
go build -ldflags="-s -w" -o Discord-WebSetup.exe ./cmd/websetup
```

---

## 📄 License
MIT License
