/**
 * i18n Translation Dictionary for Discord PTB Portable + Vencord WebSetup
 * Note: ZERO EMOJIS - Strictly clean professional text only.
 */

const translations = {
  th: {
    // Header & Meta
    appTitle: "Discord PTB Portable",
    appSubtitle: "ระบบติดตั้งแบบพกพา พร้อม Vencord และความปลอดภัยขั้นสูง",
    badgeSecure: "ระบบความปลอดภัยป้องกันการดักจับข้อมูล",
    badgeSelfUpdating: "รองรับการอัปเดตอัตโนมัติ",
    langTh: "ไทย",
    langEn: "EN",

    // Step Indicator
    step1Title: "โหมดติดตั้ง",
    step2Title: "เลือกโปรไฟล์",
    step3Title: "กำลังติดตั้ง",
    step4Title: "เสร็จสิ้น",

    // Step 1: Mode Selection
    modeSectionTitle: "เลือกรูปแบบการติดตั้ง",
    modeSectionDesc: "เลือกรูปแบบที่เหมาะสมกับการใช้งานของคุณมากที่สุด",
    modePcTitle: "ติดตั้งบนคอมพิวเตอร์ (Standard PC)",
    modePcDesc: "เหมาะสำหรับการใช้งานประจำเครื่อง สร้างไอคอนทางลัดบน Desktop และเชื่อมต่อ Rich Presence สมบูรณ์แบบ",
    modeUsbTitle: "พกพาเต็มรูปแบบ (True USB Portable)",
    modeUsbDesc: "ไม่เขียนข้อมูลลง Windows Registry รันจาก Flash Drive ได้ทันทีโดยไม่ทิ้งร่องรอยในเครื่อง",
    installPathLabel: "โฟลเดอร์สำหรับติดตั้ง",
    browseFolderBtn: "เลือกโฟลเดอร์",
    nextBtn: "ถัดไป",

    // Step 2: Profiles & Pro Features
    profileSectionTitle: "เลือกโปรไฟล์สี Discord (1-12 บัญชี)",
    profileSectionDesc: "คลิกเลือกสีประจำตัวสำหรับแต่ละบัญชี ระบบจะสร้างไอคอนสีและแยกพื้นที่จัดเก็บข้อมูลให้โดยอัตโนมัติ",
    selectAllBtn: "เลือกทั้งหมด",
    deselectAllBtn: "ล้างการเลือก",
    selectedCount: "เลือกแล้ว {count} โปรไฟล์",
    primaryRpcBadge: "Primary RPC (Pipe 0)",
    secondaryRpcBadge: "Isolated Pipe",

    // Pro Features
    proFeaturesTitle: "ฟีเจอร์ระดับโปร (Pro Settings)",
    ecoModeTitle: "Eco / Low-RAM Mode (โหมดประหยัดแรม)",
    ecoModeDesc: "ลดอัตราเรนเดอร์เฟรมและจำกัดการทำงานเบื้องหลัง ประหยัดแรมสูงสุด 40-60%",
    securityShieldTitle: "Anti-Grabber Security Shield",
    securityShieldDesc: "ป้องกันการดักจับ Username, Password และ Token โดยจำกัดสิทธิ์ NTFS และบล็อก Debugging Flags",
    vencordTitle: "ติดตั้ง Vencord Client Mod",
    vencordDesc: "ติดตั้งปลั๊กอินและธีมตกแต่ง ปรับแต่งการใช้งาน Discord ได้อย่างอิสระ",
    backupTitle: "สำรองและกู้คืนข้อมูลโปรไฟล์",
    backupDesc: "จัดเก็บหรือนำเข้าข้อมูลโปรไฟล์ทั้งหมดในรูปแบบไฟล์บีบอัด (.zip)",
    backupBtn: "สำรองข้อมูล (Backup)",
    restoreBtn: "กู้คืนข้อมูล (Restore)",
    backBtn: "ย้อนกลับ",
    startInstallBtn: "เริ่มการติดตั้ง",

    // Step 3: Installation Progress
    installingTitle: "กำลังดำเนินการติดตั้ง",
    installingDesc: "ระบบกำลังเตรียมไฟล์ Discord PTB, Vencord และคอนฟิกโปรไฟล์ความปลอดภัย...",
    stage1Name: "ดาวน์โหลดแพ็กเกจ Discord PTB ล่าสุด",
    stage2Name: "ผสานระบบ Vencord Client Mod",
    stage3Name: "จัดเตรียมโฟลเดอร์โปรไฟล์และเกราะความปลอดภัย",
    stage4Name: "สร้างทางลัดบน Desktop และไอคอนสีประจำถาดระบบ",
    statusPreparing: "กำลังเตรียมระบบ...",
    statusDownloading: "กำลังดาวน์โหลดไฟล์ Discord PTB...",
    statusVencord: "กำลังติดตั้งและตั้งค่า Vencord...",
    statusHardening: "กำลังเปิดใช้งานเกราะป้องกัน Anti-Grabber และสิทธิ์ NTFS...",
    statusShortcuts: "กำลังสร้างไอคอนทางลัดสีสันสดใส...",
    statusCompleted: "การติดตั้งเสร็จสมบูรณ์เรียบร้อยแล้ว",
    viewLogsToggle: "แสดงบันทึกการทำงาน (Logs)",

    // Step 4: Finished
    finishTitle: "ติดตั้งเสร็จสิ้นเรียบร้อยแล้ว",
    finishDesc: "Discord PTB Portable พร้อมใช้งานสำหรับบัญชีทั้งหมดที่คุณเลือกแล้ว",
    summaryModeLabel: "รูปแบบการติดตั้ง:",
    summaryProfilesLabel: "จำนวนโปรไฟล์ที่ติดตั้ง:",
    summarySecurityLabel: "ระบบความปลอดภัย:",
    summarySecurityActive: "เปิดใช้งาน (ป้องกัน Token สมบูรณ์แบบ)",
    launchNowLabel: "เปิด Discord PTB ทันทีหลังปิดตัวติดตั้ง",
    launchProfileSelectLabel: "เลือกโปรไฟล์ที่จะเปิด:",
    createDesktopShortcutsLabel: "สร้างไอคอนทางลัดบนหน้าจอเดสก์ท็อป",
    finishBtn: "เสร็จสิ้นและเปิดโปรแกรม",
    finishOnlyBtn: "เสร็จสิ้น",
    openFolderBtn: "เปิดโฟลเดอร์ติดตั้ง",

    // Alerts and Dialogs
    alertSelectAtLeastOne: "กรุณาเลือกโปรไฟล์อย่างน้อย 1 สีเพื่อดำเนินการต่อ",
    alertBackupSuccess: "สำรองข้อมูลโปรไฟล์สำเร็จเรียบร้อยแล้ว",
    alertRestoreSuccess: "กู้คืนข้อมูลโปรไฟล์สำเร็จเรียบร้อยแล้ว",
    alertError: "เกิดข้อผิดพลาด: ",
  },

  en: {
    // Header & Meta
    appTitle: "Discord PTB Portable",
    appSubtitle: "Portable Setup with Vencord & Advanced Security Protection",
    badgeSecure: "Anti-Grabber Security Guard Active",
    badgeSelfUpdating: "Self-Updating Ready",
    langTh: "TH",
    langEn: "English",

    // Step Indicator
    step1Title: "Install Mode",
    step2Title: "Profiles & Colors",
    step3Title: "Installation",
    step4Title: "Complete",

    // Step 1: Mode Selection
    modeSectionTitle: "Select Installation Mode",
    modeSectionDesc: "Choose the deployment style that best fits your workflow",
    modePcTitle: "Standard PC Installation",
    modePcDesc: "Best for daily desktop usage. Generates color-coded desktop shortcuts and seamless Rich Presence.",
    modeUsbTitle: "True USB Portable Mode",
    modeUsbDesc: "Zero Windows Registry writes. Run cleanly from USB flash drives without leaving traces behind.",
    installPathLabel: "Target Installation Directory",
    browseFolderBtn: "Browse",
    nextBtn: "Next",

    // Step 2: Profiles & Pro Features
    profileSectionTitle: "Select Colored Discord Profiles (1-12 Accounts)",
    profileSectionDesc: "Click to select distinct color themes for each account. Independent data isolation and colored icons are configured automatically.",
    selectAllBtn: "Select All",
    deselectAllBtn: "Clear Selection",
    selectedCount: "{count} Profiles Selected",
    primaryRpcBadge: "Primary RPC (Pipe 0)",
    secondaryRpcBadge: "Isolated Pipe",

    // Pro Features
    proFeaturesTitle: "Professional Settings",
    ecoModeTitle: "Eco / Low-RAM Mode",
    ecoModeDesc: "Throttles background rendering frames and trims resource footprint by up to 40-60%.",
    securityShieldTitle: "Anti-Grabber Security Shield",
    securityShieldDesc: "Hardens credentials against stealers by applying restricted NTFS ACLs and filtering inspection flags.",
    vencordTitle: "Install Vencord Client Mod",
    vencordDesc: "Empower Discord with customizable plugins, performance patches, and user themes.",
    backupTitle: "Profile Backup & Recovery",
    backupDesc: "Export or restore all user profiles into a compressed archive (.zip).",
    backupBtn: "Backup Profiles",
    restoreBtn: "Restore Profiles",
    backBtn: "Back",
    startInstallBtn: "Start Installation",

    // Step 3: Installation Progress
    installingTitle: "Installing Discord PTB Suite",
    installingDesc: "Deploying Discord PTB core, Vencord mod, and isolated security barriers...",
    stage1Name: "Fetch Latest Discord PTB Core Package",
    stage2Name: "Inject Vencord Client Mod",
    stage3Name: "Configure Profile Sandboxes & Security Shield",
    stage4Name: "Generate Color Desktop Shortcuts & System Tray Assets",
    statusPreparing: "Preparing workspace...",
    statusDownloading: "Downloading Discord PTB binary assets...",
    statusVencord: "Configuring Vencord mod loader...",
    statusHardening: "Enforcing Anti-Grabber guard and NTFS security descriptors...",
    statusShortcuts: "Generating color-coded desktop shortcuts...",
    statusCompleted: "Installation finished successfully.",
    viewLogsToggle: "View detailed logs",

    // Step 4: Finished
    finishTitle: "Installation Complete",
    finishDesc: "Discord PTB Portable is ready for all your selected colored accounts.",
    summaryModeLabel: "Deployment Mode:",
    summaryProfilesLabel: "Installed Profiles:",
    summarySecurityLabel: "Security Defense:",
    summarySecurityActive: "Active (Hardened Token Protection)",
    launchNowLabel: "Launch Discord PTB immediately on exit",
    launchProfileSelectLabel: "Launch Profile:",
    createDesktopShortcutsLabel: "Create desktop shortcuts for selected profiles",
    finishBtn: "Finish & Launch",
    finishOnlyBtn: "Finish",
    openFolderBtn: "Open Target Folder",

    // Alerts and Dialogs
    alertSelectAtLeastOne: "Please select at least 1 color profile to proceed.",
    alertBackupSuccess: "Profile backup completed successfully.",
    alertRestoreSuccess: "Profile restoration completed successfully.",
    alertError: "Error encountered: ",
  }
};

window.I18N = {
  currentLang: "th",

  setLanguage(lang) {
    if (translations[lang]) {
      this.currentLang = lang;
      this.applyTranslations();
      localStorage.setItem("preferred_lang", lang);
    }
  },

  t(key, params = {}) {
    const dict = translations[this.currentLang] || translations.th;
    let text = dict[key] || translations.en[key] || key;
    for (const [pKey, pVal] of Object.entries(params)) {
      text = text.replace(new RegExp(`\\{${pKey}\\}`, "g"), pVal);
    }
    return text;
  },

  applyTranslations() {
    // Translate all elements with data-i18n attribute
    document.querySelectorAll("[data-i18n]").forEach((el) => {
      const key = el.getAttribute("data-i18n");
      const translation = this.t(key);
      if (translation) {
        el.textContent = translation;
      }
    });

    // Translate placeholder attributes with data-i18n-placeholder
    document.querySelectorAll("[data-i18n-placeholder]").forEach((el) => {
      const key = el.getAttribute("data-i18n-placeholder");
      const translation = this.t(key);
      if (translation) {
        el.setAttribute("placeholder", translation);
      }
    });

    // Update active state on language buttons
    const btnTh = document.getElementById("lang-btn-th");
    const btnEn = document.getElementById("lang-btn-en");
    if (btnTh && btnEn) {
      if (this.currentLang === "th") {
        btnTh.classList.add("active");
        btnEn.classList.remove("active");
      } else {
        btnEn.classList.add("active");
        btnTh.classList.remove("active");
      }
    }

    // Dispatch event for components that need dynamic re-rendering
    window.dispatchEvent(new CustomEvent("languageChanged", { detail: { lang: this.currentLang } }));
  },

  init() {
    const saved = localStorage.getItem("preferred_lang");
    if (saved && (saved === "th" || saved === "en")) {
      this.currentLang = saved;
    } else {
      // Default to Thai as per requirements
      this.currentLang = "th";
    }
    this.applyTranslations();
  }
};
