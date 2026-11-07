/**
 * i18n Translation Dictionary for Discord PTB Portable + Vencord WebSetup
 * Strict Constraint: ZERO EMOJIS - Strictly clean professional text only.
 * Principles: 10 Usability Heuristics & Cognitive Clarity
 */

const translations = {
  th: {
    // Header & Meta
    appTitle: "Discord PTB Portable",
    langTh: "ไทย",
    langEn: "EN",
    winMinimize: "ย่อหน้าต่าง",
    winMaximize: "ขยายหน้าต่าง",
    winClose: "ปิด",

    // Step Indicator
    step1Title: "โฟลเดอร์ติดตั้ง",
    step2Title: "เลือกโปรไฟล์",
    step3Title: "กำลังติดตั้ง",
    step4Title: "เสร็จสิ้น",

    // Step 1: Install Path
    installPathLabel: "โฟลเดอร์สำหรับติดตั้ง",
    installPathDesc: "ระบุตำแหน่งที่ต้องการติดตั้ง Discord PTB Portable บนเครื่องของคุณ",
    browseFolderBtn: "เลือกโฟลเดอร์",
    nextBtn: "ถัดไป",

    // Step 2: Profiles & Vencord
    profileSectionTitle: "เลือกโปรไฟล์สี Discord (1-12 บัญชี)",
    profileSectionDesc: "คลิกเลือกสีประจำตัวสำหรับแต่ละบัญชี ระบบจะสร้างไอคอนสีและแยกพื้นที่จัดเก็บข้อมูลให้โดยอัตโนมัติ",
    selectAllBtn: "เลือกทั้งหมด",
    deselectAllBtn: "ล้างการเลือก",
    selectedCount: "เลือกแล้ว {count} โปรไฟล์",
    vencordTitle: "ติดตั้ง Vencord Client Mod",
    vencordDesc: "ติดตั้งปลั๊กอินและธีมตกแต่ง ปรับแต่งการใช้งาน Discord ได้อย่างอิสระ",
    launchNowLabel: "เปิด Discord PTB ทันทีหลังติดตั้งเสร็จ",
    launchNowDesc: "เปิดโปรแกรมโปรไฟล์แรกทันทีเมื่อกดเสร็จสิ้น",
    createDesktopShortcutsLabel: "สร้างไอคอนทางลัดไว้หน้าเดสก์ท็อป",
    createDesktopShortcutsDesc: "สร้างทางลัดพร้อมไอคอนสีของแต่ละบัญชีบนหน้าจอ Desktop",
    backBtn: "ย้อนกลับ",
    startInstallBtn: "เริ่มการติดตั้ง",

    // Step 3: Installation Progress
    installingTitle: "กำลังดำเนินการติดตั้ง",
    installingDesc: "ระบบกำลังเตรียมไฟล์ Discord PTB, Vencord และพื้นที่โปรไฟล์อย่างปลอดภัย...",
    stage1Name: "ดาวน์โหลดแพ็กเกจ Discord PTB ล่าสุด",
    stage2Name: "ผสานระบบ Vencord Client Mod",
    stage3Name: "จัดเตรียมโฟลเดอร์โปรไฟล์และความปลอดภัย",
    stage4Name: "สร้างทางลัดบน Desktop และไอคอนสีประจำระบบ",
    statusPreparing: "กำลังเตรียมระบบ...",
    statusDownloading: "กำลังดาวน์โหลดไฟล์ Discord PTB...",
    statusVencord: "กำลังติดตั้งและตั้งค่า Vencord...",
    statusHardening: "กำลังเปิดใช้งานเกราะป้องกัน Anti-Grabber และสิทธิ์ NTFS...",
    statusShortcuts: "กำลังสร้างไอคอนทางลัดสีสันสดใส...",
    statusCompleted: "การติดตั้งเสร็จสมบูรณ์เรียบร้อยแล้ว",

    // Step 4: Finished
    finishTitle: "ติดตั้งเสร็จสิ้นเรียบร้อยแล้ว",
    finishDesc: "Discord PTB Portable พร้อมใช้งานสำหรับบัญชีทั้งหมดที่คุณเลือกแล้ว",
    finishBtn: "เสร็จสิ้น",

    // Alerts
    alertSelectAtLeastOne: "กรุณาเลือกโปรไฟล์อย่างน้อย 1 สีเพื่อดำเนินการต่อ",
    alertError: "เกิดข้อผิดพลาด: ",
  },

  en: {
    // Header & Meta
    appTitle: "Discord PTB Portable",
    langTh: "TH",
    langEn: "English",
    winMinimize: "Minimize",
    winMaximize: "Maximize",
    winClose: "Close",

    // Step Indicator
    step1Title: "Install Path",
    step2Title: "Select Profiles",
    step3Title: "Installing",
    step4Title: "Completed",

    // Step 1: Install Path
    installPathLabel: "Installation Folder",
    installPathDesc: "Specify the directory where Discord PTB Portable will be installed on your system",
    browseFolderBtn: "Browse Folder",
    nextBtn: "Next",

    // Step 2: Profiles & Vencord
    profileSectionTitle: "Select Discord Colored Profiles (1-12 Accounts)",
    profileSectionDesc: "Click to select color profiles for each account. The system creates colored icons and isolated data storage.",
    selectAllBtn: "Select All",
    deselectAllBtn: "Clear Selection",
    selectedCount: "{count} profile(s) selected",
    vencordTitle: "Install Vencord Client Mod",
    vencordDesc: "Install client plugins and custom themes to personalize your Discord experience.",
    launchNowLabel: "Launch Discord PTB immediately after install",
    launchNowDesc: "Open primary profile window when clicking Finish",
    createDesktopShortcutsLabel: "Create desktop shortcuts with colored icons",
    createDesktopShortcutsDesc: "Place colored shortcuts on desktop for each profile",
    backBtn: "Back",
    startInstallBtn: "Install Now",

    // Step 3: Installation Progress
    installingTitle: "Installation in Progress",
    installingDesc: "Setting up Discord PTB, Vencord, and secure profile directories...",
    stage1Name: "Download latest Discord PTB package",
    stage2Name: "Inject Vencord Client Mod",
    stage3Name: "Prepare isolated profiles and security shield",
    stage4Name: "Create desktop shortcuts with colored icons",
    statusPreparing: "Preparing setup...",
    statusDownloading: "Downloading Discord PTB package...",
    statusVencord: "Configuring Vencord Client Mod...",
    statusHardening: "Enforcing Anti-Grabber shield and NTFS DACL...",
    statusShortcuts: "Creating desktop shortcuts...",
    statusCompleted: "Installation completed successfully",

    // Step 4: Finished
    finishTitle: "Installation Completed Successfully",
    finishDesc: "Discord PTB Portable is ready to use for all selected accounts.",
    finishBtn: "Finish",

    // Alerts
    alertSelectAtLeastOne: "Please select at least 1 color profile to proceed.",
    alertError: "An error occurred: ",
  },
};

class I18nManager {
  constructor() {
    this.currentLang = localStorage.getItem("app_lang") || "th";
  }

  init() {
    this.applyLanguage(this.currentLang);
  }

  setLanguage(lang) {
    if (translations[lang]) {
      this.currentLang = lang;
      localStorage.setItem("app_lang", lang);
      this.applyLanguage(lang);
      document.documentElement.lang = lang;
    }
  }

  t(key, params = {}) {
    let str = translations[this.currentLang]?.[key] || translations["th"]?.[key] || key;
    for (const [param, val] of Object.entries(params)) {
      str = str.replace(`{${param}}`, val);
    }
    return str;
  }

  applyLanguage(lang) {
    document.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.getAttribute("data-i18n");
      const translation = this.t(key);
      if (translation) {
        if (el.tagName === "INPUT" && el.type === "button") {
          el.value = translation;
        } else {
          el.textContent = translation;
        }
      }
    });

    document.querySelectorAll("[data-i18n-title]").forEach((el) => {
      const key = el.getAttribute("data-i18n-title");
      const translation = this.t(key);
      if (translation) {
        el.title = translation;
      }
    });
  }
}

window.I18N = new I18nManager();
