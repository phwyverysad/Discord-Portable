/**
 * i18n Translation Dictionary for Discord PTB Portable + Vencord WebSetup
 * Note: ZERO EMOJIS - Strictly clean professional text only.
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
    launchNowLabel: "เปิด Discord PTB ทันทีหลังปิดตัวติดตั้ง",
    createDesktopShortcutsLabel: "สร้างไอคอนทางลัดบนหน้าจอเดสก์ท็อป",
    finishBtn: "เสร็จสิ้นและเปิดโปรแกรม",

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
    launchNowLabel: "Launch Discord PTB now",
    createDesktopShortcutsLabel: "Create desktop shortcuts with colored icons",
    finishBtn: "Finish and Launch",

    // Alerts
    alertSelectAtLeastOne: "Please select at least 1 color profile to proceed.",
    alertError: "An error occurred: ",
  },
};

class I18nManager {
  constructor() {
    this.currentLang = "th";
  }

  init() {
    this.applyLanguage();
  }

  setLanguage(lang) {
    if (!translations[lang]) return;
    this.currentLang = lang;
    this.applyLanguage();

    const btnTh = document.getElementById("lang-btn-th");
    const btnEn = document.getElementById("lang-btn-en");
    if (btnTh && btnEn) {
      if (lang === "th") {
        btnTh.classList.add("active");
        btnEn.classList.remove("active");
      } else {
        btnEn.classList.add("active");
        btnTh.classList.remove("active");
      }
    }

    window.dispatchEvent(new CustomEvent("languageChanged", { detail: { lang } }));
  }

  t(key, replacements = {}) {
    const dict = translations[this.currentLang] || translations.th;
    let text = dict[key] || translations.th[key] || key;

    Object.keys(replacements).forEach((k) => {
      text = text.replace(new RegExp(`\\{${k}\\}`, "g"), replacements[k]);
    });

    return text;
  }

  applyLanguage() {
    document.documentElement.lang = this.currentLang;
    const elements = document.querySelectorAll("[data-i18n]");

    elements.forEach((el) => {
      const key = el.getAttribute("data-i18n");
      const translation = this.t(key);
      if (el.tagName === "INPUT" && el.getAttribute("placeholder")) {
        el.setAttribute("placeholder", translation);
      } else {
        el.textContent = translation;
      }
    });
  }
}

window.I18N = new I18nManager();
