/**
 * i18n Translation Dictionary for Discord PTB Portable + Vencord WebSetup
 * Strict Constraint: ZERO EMOJIS - Strictly clean professional text only.
 * Principles: Concise UX Copywriting, 10 Usability Heuristics & Cognitive Clarity
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
    installPathLabel: "โฟลเดอร์ติดตั้ง",
    installPathDesc: "เลือกตำแหน่งจัดเก็บโปรแกรมบนเครื่องของคุณ",
    installPathCallout: "แยกโปรไฟล์และข้อมูลอิสระ ไม่ปะปนกับ Discord หลัก",
    browseFolderBtn: "เลือกโฟลเดอร์",
    nextBtn: "ถัดไป",

    // Step 2: Profiles & Vencord
    profileSectionTitle: "เลือกโปรไฟล์สี",
    profileSectionDesc: "เลือกสีประจำบัญชีเพื่อแยกไอคอนและข้อมูล",
    selectAllBtn: "เลือกทั้งหมด",
    deselectAllBtn: "ล้างการเลือก",
    selectedCount: "เลือกแล้ว {count} บัญชี",
    vencordTitle: "ติดตั้ง Vencord",
    vencordDesc: "เพิ่มปลั๊กอินและธีมปรับแต่ง",
    launchNowLabel: "เปิด Discord ทันที",
    launchNowDesc: "เปิดใช้งานทันทีหลังติดตั้งเสร็จ",
    createDesktopShortcutsLabel: "สร้างทางลัดบนเดสก์ท็อป",
    createDesktopShortcutsDesc: "สร้างไอคอนสีแยกแต่ละโปรไฟล์บน Desktop",
    backBtn: "ย้อนกลับ",
    startInstallBtn: "ติดตั้งทันที",

    // Step 3: Installation Progress
    installingTitle: "กำลังติดตั้ง",
    installingDesc: "ระบบกำลังจัดเตรียมไฟล์และโปรไฟล์...",
    stage1Name: "ดาวน์โหลด Discord PTB",
    stage2Name: "ติดตั้ง Vencord",
    stage3Name: "จัดเตรียมพื้นที่โปรไฟล์",
    stage4Name: "สร้างทางลัดและไอคอนสี",
    statusPreparing: "กำลังเตรียมระบบ...",
    statusDownloading: "กำลังดาวน์โหลดไฟล์...",
    statusVencord: "กำลังติดตั้ง Vencord...",
    statusHardening: "กำลังตั้งค่าความปลอดภัย...",
    statusShortcuts: "กำลังสร้างไอคอนทางลัด...",
    statusCompleted: "ติดตั้งเสร็จสมบูรณ์",

    // Step 4: Finished
    finishTitle: "ติดตั้งเสร็จสมบูรณ์",
    finishDesc: "Discord PTB พร้อมใช้งานแล้ว",
    finishBtn: "เสร็จสิ้น",

    // Alerts
    alertSelectAtLeastOne: "กรุณาเลือกอย่างน้อย 1 สี",
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
    step2Title: "Profiles",
    step3Title: "Installing",
    step4Title: "Completed",

    // Step 1: Install Path
    installPathLabel: "Installation Folder",
    installPathDesc: "Choose where to install Discord PTB Portable",
    installPathCallout: "Isolated data and cache. Independent from system Discord.",
    browseFolderBtn: "Browse",
    nextBtn: "Next",

    // Step 2: Profiles & Vencord
    profileSectionTitle: "Select Color Profiles",
    profileSectionDesc: "Choose colors to separate icons and account data",
    selectAllBtn: "Select All",
    deselectAllBtn: "Clear",
    selectedCount: "{count} selected",
    vencordTitle: "Install Vencord",
    vencordDesc: "Add plugins and custom themes",
    launchNowLabel: "Launch Discord immediately",
    launchNowDesc: "Open profile window when finished",
    createDesktopShortcutsLabel: "Create desktop shortcuts",
    createDesktopShortcutsDesc: "Place colored icons on desktop",
    backBtn: "Back",
    startInstallBtn: "Install Now",

    // Step 3: Installation Progress
    installingTitle: "Installing",
    installingDesc: "Setting up files and profile data...",
    stage1Name: "Download Discord PTB",
    stage2Name: "Install Vencord",
    stage3Name: "Setup profile data",
    stage4Name: "Create shortcuts and icons",
    statusPreparing: "Preparing setup...",
    statusDownloading: "Downloading files...",
    statusVencord: "Installing Vencord...",
    statusHardening: "Setting up security...",
    statusShortcuts: "Creating shortcuts...",
    statusCompleted: "Setup complete",

    // Step 4: Finished
    finishTitle: "Setup Complete",
    finishDesc: "Discord PTB is ready to use.",
    finishBtn: "Finish",

    // Alerts
    alertSelectAtLeastOne: "Please select at least 1 color profile.",
    alertError: "Error: ",
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
    const dict = translations[lang] || translations["th"];
    document.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.getAttribute("data-i18n");
      if (dict[key]) {
        if (el.tagName === "INPUT" && el.getAttribute("type") === "button") {
          el.value = dict[key];
        } else {
          el.textContent = dict[key];
        }
      }
    });

    document.querySelectorAll("[data-i18n-title]").forEach((el) => {
      const key = el.getAttribute("data-i18n-title");
      if (dict[key]) {
        el.setAttribute("title", dict[key]);
      }
    });
  }
}

window.I18N = new I18nManager();
