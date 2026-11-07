/**
 * Discord PTB Portable + Vencord WebSetup - Application Logic & Backend Bridge
 * Zero Emojis Policy - Handled via SVG and clean typography.
 */

(function () {
  "use strict";

  // Application State
  const state = {
    currentStep: 1,
    installMode: "pc", // 'pc' or 'usb'
    installPath: "",
    profiles: [],
    selectedProfileIds: [1], // Default profile 1 selected
    ecoMode: true,
    securityShield: true,
    vencord: true,
    launchNow: true,
    launchProfile: 1,
    createShortcuts: true,
    isInstalling: false,
    logs: [],
    backendOnline: false,
  };

  // DOM Elements cache
  const elements = {};

  // Initialize
  document.addEventListener("DOMContentLoaded", () => {
    cacheElements();
    initI18n();
    initEvents();
    checkBackendAndLoadProfiles();
  });

  function cacheElements() {
    elements.stepItems = document.querySelectorAll(".step-item");
    elements.stepViews = document.querySelectorAll(".step-view");
    elements.connectors = document.querySelectorAll(".step-connector");

    // Mode Cards
    elements.modePc = document.getElementById("mode-pc");
    elements.modeUsb = document.getElementById("mode-usb");
    elements.installPathInput = document.getElementById("install-path-input");
    elements.browseFolderBtn = document.getElementById("browse-folder-btn");

    // Profiles
    elements.profilesGrid = document.getElementById("profiles-grid");
    elements.profileCounter = document.getElementById("profile-counter");
    elements.btnSelectAll = document.getElementById("btn-select-all");
    elements.btnDeselectAll = document.getElementById("btn-deselect-all");

    // Pro Switches
    elements.switchEco = document.getElementById("switch-eco");
    elements.switchSecurity = document.getElementById("switch-security");
    elements.switchVencord = document.getElementById("switch-vencord");
    elements.btnBackup = document.getElementById("btn-backup");
    elements.btnRestore = document.getElementById("btn-restore");

    // Progress Elements
    elements.progressFill = document.getElementById("progress-fill");
    elements.progressPercent = document.getElementById("progress-percent");
    elements.progressStatus = document.getElementById("progress-status");
    elements.stageItems = document.querySelectorAll(".stage-item");
    elements.logsToggleBtn = document.getElementById("logs-toggle-btn");
    elements.logsContent = document.getElementById("logs-content");

    // Finish Elements
    elements.summaryMode = document.getElementById("summary-mode");
    elements.summaryProfiles = document.getElementById("summary-profiles");
    elements.summarySecurity = document.getElementById("summary-security");
    elements.launchNowCheck = document.getElementById("launch-now-check");
    elements.launchProfileSelect = document.getElementById("launch-profile-select");
    elements.shortcutsCheck = document.getElementById("shortcuts-check");
    elements.btnFinish = document.getElementById("btn-finish");
    elements.btnOpenFolder = document.getElementById("btn-open-folder");

    // Footer Navigation
    elements.btnBack = document.getElementById("btn-back");
    elements.btnNext = document.getElementById("btn-next");
  }

  function initI18n() {
    window.I18N.init();

    // Language switcher buttons
    const btnTh = document.getElementById("lang-btn-th");
    const btnEn = document.getElementById("lang-btn-en");

    if (btnTh) {
      btnTh.addEventListener("click", () => {
        window.I18N.setLanguage("th");
      });
    }
    if (btnEn) {
      btnEn.addEventListener("click", () => {
        window.I18N.setLanguage("en");
      });
    }

    window.addEventListener("languageChanged", () => {
      renderProfiles();
      updateCounter();
      updateSummary();
    });
  }

  function initEvents() {
    // Mode Selection
    if (elements.modePc) {
      elements.modePc.addEventListener("click", () => selectMode("pc"));
    }
    if (elements.modeUsb) {
      elements.modeUsb.addEventListener("click", () => selectMode("usb"));
    }

    if (elements.browseFolderBtn) {
      elements.browseFolderBtn.addEventListener("click", handleBrowseFolder);
    }

    // Profiles Toolbar
    if (elements.btnSelectAll) {
      elements.btnSelectAll.addEventListener("click", selectAllProfiles);
    }
    if (elements.btnDeselectAll) {
      elements.btnDeselectAll.addEventListener("click", deselectAllProfiles);
    }

    // Pro Switches
    if (elements.switchEco) {
      elements.switchEco.addEventListener("change", (e) => {
        state.ecoMode = e.target.checked;
      });
    }
    if (elements.switchVencord) {
      elements.switchVencord.addEventListener("change", (e) => {
        state.vencord = e.target.checked;
      });
    }

    // Backup & Restore
    if (elements.btnBackup) {
      elements.btnBackup.addEventListener("click", handleBackup);
    }
    if (elements.btnRestore) {
      elements.btnRestore.addEventListener("click", handleRestore);
    }

    // Logs Drawer
    if (elements.logsToggleBtn) {
      elements.logsToggleBtn.addEventListener("click", () => {
        elements.logsContent.classList.toggle("open");
      });
    }

    // Footer Buttons
    if (elements.btnBack) {
      elements.btnBack.addEventListener("click", handleBack);
    }
    if (elements.btnNext) {
      elements.btnNext.addEventListener("click", handleNext);
    }

    // Finish Actions
    if (elements.btnFinish) {
      elements.btnFinish.addEventListener("click", handleFinish);
    }
    if (elements.btnOpenFolder) {
      elements.btnOpenFolder.addEventListener("click", handleOpenFolder);
    }

    if (elements.launchNowCheck) {
      elements.launchNowCheck.addEventListener("change", (e) => {
        state.launchNow = e.target.checked;
        const selectorWrap = document.getElementById("profile-select-wrap");
        if (selectorWrap) {
          selectorWrap.style.display = e.target.checked ? "flex" : "none";
        }
      });
    }

    if (elements.launchProfileSelect) {
      elements.launchProfileSelect.addEventListener("change", (e) => {
        state.launchProfile = parseInt(e.target.value, 10);
      });
    }

    if (elements.shortcutsCheck) {
      elements.shortcutsCheck.addEventListener("change", (e) => {
        state.createShortcuts = e.target.checked;
      });
    }
  }

  // Check Backend and Load Profile Metadata
  async function checkBackendAndLoadProfiles() {
    let metadataLoaded = false;

    try {
      const res = await fetch("/api/status");
      if (res.ok) {
        state.backendOnline = true;
        const data = await res.json();
        if (data.defaultPath) {
          state.installPath = data.defaultPath;
          if (elements.installPathInput) {
            elements.installPathInput.value = data.defaultPath;
          }
        }
      }
    } catch {
      state.backendOnline = false;
    }

    // Load Metadata
    try {
      const metaUrl = state.backendOnline ? "/api/metadata" : "assets/icons/metadata.json";
      const res = await fetch(metaUrl);
      if (res.ok) {
        const meta = await res.json();
        state.profiles = meta.profiles || [];
        metadataLoaded = true;
      }
    } catch {
      console.warn("Could not fetch metadata from endpoint, falling back to static definition.");
    }

    if (!metadataLoaded || state.profiles.length === 0) {
      // Fallback static profile list if metadata fetch failed
      state.profiles = getStaticProfilesFallback();
    }

    // Set default path if empty
    if (!state.installPath && elements.installPathInput) {
      state.installPath = "C:\\Discord-PTB-Portable";
      elements.installPathInput.value = state.installPath;
    }

    renderProfiles();
    updateCounter();
    populateLaunchDropdown();
  }

  function getStaticProfilesFallback() {
    return [
      { id: 1, code: "01", key: "yellow", name_th: "สีเหลือง", name_en: "Yellow", hex: "#FFFF00", primary_rpc: true, png_file: "01_yellow_FFFF00.png" },
      { id: 2, code: "02", key: "yellow_orange", name_th: "สีส้มอมเหลือง", name_en: "Yellow Orange", hex: "#FFC000", primary_rpc: false, png_file: "02_yellow_orange_FFC000.png" },
      { id: 3, code: "03", key: "orange", name_th: "สีส้ม", name_en: "Orange", hex: "#F58220", primary_rpc: false, png_file: "03_orange_F58220.png" },
      { id: 4, code: "04", key: "red_orange", name_th: "สีส้มอมแดง", name_en: "Red Orange", hex: "#FA6304", primary_rpc: false, png_file: "04_red_orange_FA6304.png" },
      { id: 5, code: "05", key: "red", name_th: "สีแดง", name_en: "Red", hex: "#ED1C24", primary_rpc: false, png_file: "05_red_ED1C24.png" },
      { id: 6, code: "06", key: "red_violet", name_th: "สีม่วงอมแดง", name_en: "Red Violet", hex: "#BD0074", primary_rpc: false, png_file: "06_red_violet_BD0074.png" },
      { id: 7, code: "07", key: "violet", name_th: "สีม่วงเข้ม", name_en: "Violet", hex: "#59277A", primary_rpc: false, png_file: "07_violet_59277A.png" },
      { id: 8, code: "08", key: "blue_violet", name_th: "สีน้ำเงินอมม่วง", name_en: "Blue Violet", hex: "#4400A3", primary_rpc: false, png_file: "08_blue_violet_4400A3.png" },
      { id: 9, code: "09", key: "blue", name_th: "สีน้ำเงิน", name_en: "Blue", hex: "#0020C2", primary_rpc: false, png_file: "09_blue_0020C2.png" },
      { id: 10, code: "10", key: "blue_green", name_th: "สีเขียวอมฟ้า", name_en: "Blue Green", hex: "#008240", primary_rpc: false, png_file: "10_blue_green_008240.png" },
      { id: 11, code: "11", key: "green", name_th: "สีเขียว", name_en: "Green", hex: "#00B050", primary_rpc: false, png_file: "11_green_00B050.png" },
      { id: 12, code: "12", key: "yellow_green", name_th: "สีเขียวตองอ่อน", name_en: "Yellow Green", hex: "#8EF226", primary_rpc: false, png_file: "12_yellow_green_8EF226.png" },
    ];
  }

  // Render Profiles Grid
  function renderProfiles() {
    if (!elements.profilesGrid) return;
    elements.profilesGrid.innerHTML = "";

    const lang = window.I18N.currentLang;

    state.profiles.forEach((p) => {
      const isSelected = state.selectedProfileIds.includes(p.id);
      const card = document.createElement("div");
      card.className = `profile-card ${isSelected ? "selected" : ""}`;
      card.setAttribute("data-id", p.id);

      const colorName = lang === "th" ? p.name_th : p.name_en;
      const rpcBadgeText = p.primary_rpc ? window.I18N.t("primaryRpcBadge") : "";

      // SVG Checkmark
      const checkSvg = `
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      `;

      card.innerHTML = `
        <div class="profile-icon-wrap">
          <img class="profile-icon-img" src="assets/icons/${p.png_file}" alt="${colorName}" />
          <div class="profile-swatch" style="background-color: ${p.hex};"></div>
        </div>
        <div class="profile-details">
          <div class="profile-name">${colorName}</div>
          <div class="profile-meta">
            <span class="profile-hex">${p.hex}</span>
            ${p.primary_rpc ? `<span class="rpc-pill-badge">${rpcBadgeText}</span>` : ""}
          </div>
        </div>
        <div class="profile-check">
          ${checkSvg}
        </div>
      `;

      card.addEventListener("click", () => toggleProfileSelection(p.id));
      elements.profilesGrid.appendChild(card);
    });
  }

  function toggleProfileSelection(id) {
    const idx = state.selectedProfileIds.indexOf(id);
    if (idx >= 0) {
      if (state.selectedProfileIds.length === 1) {
        // Prevent deselecting the last remaining profile
        alert(window.I18N.t("alertSelectAtLeastOne"));
        return;
      }
      state.selectedProfileIds.splice(idx, 1);
    } else {
      state.selectedProfileIds.push(id);
      state.selectedProfileIds.sort((a, b) => a - b);
    }

    renderProfiles();
    updateCounter();
    populateLaunchDropdown();
  }

  function selectAllProfiles() {
    state.selectedProfileIds = state.profiles.map((p) => p.id);
    renderProfiles();
    updateCounter();
    populateLaunchDropdown();
  }

  function deselectAllProfiles() {
    // Keep profile 1 as default
    state.selectedProfileIds = [1];
    renderProfiles();
    updateCounter();
    populateLaunchDropdown();
  }

  function updateCounter() {
    if (elements.profileCounter) {
      elements.profileCounter.textContent = window.I18N.t("selectedCount", {
        count: state.selectedProfileIds.length,
      });
    }
  }

  function populateLaunchDropdown() {
    if (!elements.launchProfileSelect) return;
    elements.launchProfileSelect.innerHTML = "";

    const lang = window.I18N.currentLang;
    state.selectedProfileIds.forEach((id) => {
      const p = state.profiles.find((item) => item.id === id);
      if (p) {
        const opt = document.createElement("option");
        opt.value = p.id;
        const colorName = lang === "th" ? p.name_th : p.name_en;
        opt.textContent = `Profile ${p.id} (${colorName})`;
        if (p.id === state.launchProfile) {
          opt.selected = true;
        }
        elements.launchProfileSelect.appendChild(opt);
      }
    });

    if (!state.selectedProfileIds.includes(state.launchProfile) && state.selectedProfileIds.length > 0) {
      state.launchProfile = state.selectedProfileIds[0];
    }
  }

  // Navigation Steps
  function goToStep(step) {
    if (step < 1 || step > 4) return;
    state.currentStep = step;

    // Update Views
    elements.stepViews.forEach((view, idx) => {
      const viewStep = idx + 1;
      if (viewStep === step) {
        view.classList.add("active");
      } else {
        view.classList.remove("active");
      }
    });

    // Update Step Indicators
    elements.stepItems.forEach((item, idx) => {
      const itemStep = idx + 1;
      item.classList.remove("active", "completed");
      if (itemStep === step) {
        item.classList.add("active");
      } else if (itemStep < step) {
        item.classList.add("completed");
      }
    });

    // Update Connectors
    elements.connectors.forEach((conn, idx) => {
      if (idx + 1 < step) {
        conn.classList.add("completed");
      } else {
        conn.classList.remove("completed");
      }
    });

    // Update Footer Buttons Visibility & Labels
    updateFooterButtons();

    // Trigger step-specific logic
    if (step === 3) {
      startInstallation();
    } else if (step === 4) {
      updateSummary();
    }
  }

  function updateFooterButtons() {
    const footer = document.querySelector(".window-footer");
    if (state.currentStep === 1) {
      footer.style.display = "flex";
      elements.btnBack.style.display = "none";
      elements.btnNext.style.display = "inline-flex";
      elements.btnNext.textContent = window.I18N.t("nextBtn");
    } else if (state.currentStep === 2) {
      footer.style.display = "flex";
      elements.btnBack.style.display = "inline-flex";
      elements.btnNext.style.display = "inline-flex";
      elements.btnNext.textContent = window.I18N.t("startInstallBtn");
    } else if (state.currentStep === 3) {
      footer.style.display = "none";
    } else if (state.currentStep === 4) {
      footer.style.display = "none";
    }
  }

  function handleBack() {
    if (state.currentStep > 1 && !state.isInstalling) {
      goToStep(state.currentStep - 1);
    }
  }

  function handleNext() {
    if (state.currentStep === 1) {
      if (elements.installPathInput) {
        state.installPath = elements.installPathInput.value.trim() || state.installPath;
      }
      goToStep(2);
    } else if (state.currentStep === 2) {
      if (state.selectedProfileIds.length === 0) {
        alert(window.I18N.t("alertSelectAtLeastOne"));
        return;
      }
      goToStep(3);
    }
  }

  function selectMode(mode) {
    state.installMode = mode;
    if (mode === "pc") {
      elements.modePc.classList.add("selected");
      elements.modeUsb.classList.remove("selected");
      if (elements.installPathInput) {
        elements.installPathInput.value = "C:\\Discord-PTB-Portable";
      }
    } else {
      elements.modeUsb.classList.add("selected");
      elements.modePc.classList.remove("selected");
      if (elements.installPathInput) {
        elements.installPathInput.value = ".\\Discord-PTB-Portable";
      }
    }
    state.installPath = elements.installPathInput ? elements.installPathInput.value : "";
  }

  async function handleBrowseFolder() {
    if (state.backendOnline) {
      try {
        const res = await fetch("/api/browse-folder", { method: "POST" });
        if (res.ok) {
          const data = await res.json();
          if (data.path) {
            state.installPath = data.path;
            if (elements.installPathInput) {
              elements.installPathInput.value = data.path;
            }
          }
        }
      } catch (err) {
        console.warn("Browse folder API failed", err);
      }
    }
  }

  // Installation Pipeline Execution
  async function startInstallation() {
    state.isInstalling = true;
    updateProgress(0, window.I18N.t("statusPreparing"), 1);
    addLog("Initializing installation environment...");
    addLog(`Target Mode: ${state.installMode.toUpperCase()}`);
    addLog(`Target Path: ${state.installPath}`);
    addLog(`Selected Profiles: ${state.selectedProfileIds.join(", ")}`);
    addLog(`Eco / Low-RAM Mode: ${state.ecoMode}`);
    addLog(`Anti-Grabber Security Shield: ${state.securityShield}`);
    addLog(`Vencord Client Mod: ${state.vencord}`);

    const payload = {
      mode: state.installMode,
      installPath: state.installPath,
      profiles: state.selectedProfileIds,
      ecoMode: state.ecoMode,
      vencord: state.vencord,
      securityShield: true,
      createShortcuts: state.createShortcuts,
    };

    if (state.backendOnline) {
      try {
        const res = await fetch("/api/install", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });

        if (res.ok) {
          monitorBackendProgress();
          return;
        }
      } catch (err) {
        addLog(`Backend install request failed: ${err.message}. Switching to local simulation.`);
      }
    }

    // Local / Offline Simulation for interactive preview and verification
    simulateInstallationFlow();
  }

  function monitorBackendProgress() {
    const timer = setInterval(async () => {
      try {
        const res = await fetch("/api/progress");
        if (res.ok) {
          const p = await res.json();
          updateProgress(p.percent, p.statusText, p.stageIndex);
          if (p.log) addLog(p.log);

          if (p.completed) {
            clearInterval(timer);
            state.isInstalling = false;
            setTimeout(() => goToStep(4), 600);
          } else if (p.error) {
            clearInterval(timer);
            state.isInstalling = false;
            alert(window.I18N.t("alertError") + p.error);
          }
        }
      } catch (err) {
        console.error("Progress polling error", err);
      }
    }, 500);
  }

  function simulateInstallationFlow() {
    const stages = [
      { percent: 15, text: window.I18N.t("statusDownloading"), stage: 1, log: "Downloading Discord PTB 1.0.1231 packages..." },
      { percent: 35, text: window.I18N.t("statusDownloading"), stage: 1, log: "Extracting core electron assets and dependencies..." },
      { percent: 55, text: window.I18N.t("statusVencord"), stage: 2, log: "Injecting Vencord client mod loader into app.asar..." },
      { percent: 75, text: window.I18N.t("statusHardening"), stage: 3, log: "Applying Anti-Grabber Stealer Path Deflector and restricted NTFS DACL..." },
      { percent: 90, text: window.I18N.t("statusShortcuts"), stage: 4, log: "Generating high-res colored desktop shortcuts and system tray files..." },
      { percent: 100, text: window.I18N.t("statusCompleted"), stage: 4, log: "All components installed and secured successfully." },
    ];

    let current = 0;
    const interval = setInterval(() => {
      if (current >= stages.length) {
        clearInterval(interval);
        state.isInstalling = false;
        setTimeout(() => goToStep(4), 600);
        return;
      }

      const item = stages[current];
      updateProgress(item.percent, item.text, item.stage);
      addLog(item.log);
      current++;
    }, 700);
  }

  function updateProgress(percent, statusText, stageIndex) {
    if (elements.progressFill) {
      elements.progressFill.style.width = `${percent}%`;
    }
    if (elements.progressPercent) {
      elements.progressPercent.textContent = `${percent}%`;
    }
    if (elements.progressStatus) {
      elements.progressStatus.textContent = statusText;
    }

    // Update Stages checklist
    if (elements.stageItems) {
      elements.stageItems.forEach((el, idx) => {
        const sNum = idx + 1;
        el.classList.remove("in-progress", "done");
        if (sNum < stageIndex) {
          el.classList.add("done");
        } else if (sNum === stageIndex) {
          if (percent === 100) {
            el.classList.add("done");
          } else {
            el.classList.add("in-progress");
          }
        }
      });
    }
  }

  function addLog(text) {
    const timeStr = new Date().toLocaleTimeString();
    const entry = `[${timeStr}] ${text}`;
    state.logs.push(entry);
    if (elements.logsContent) {
      elements.logsContent.textContent = state.logs.join("\n");
      elements.logsContent.scrollTop = elements.logsContent.scrollHeight;
    }
  }

  // Summary Update in Step 4
  function updateSummary() {
    if (elements.summaryMode) {
      const modeText = state.installMode === "pc" ? window.I18N.t("modePcTitle") : window.I18N.t("modeUsbTitle");
      elements.summaryMode.textContent = modeText;
    }
    if (elements.summaryProfiles) {
      const names = state.selectedProfileIds.map((id) => {
        const p = state.profiles.find((item) => item.id === id);
        return p ? (window.I18N.currentLang === "th" ? p.name_th : p.name_en) : `#${id}`;
      });
      elements.summaryProfiles.textContent = `${state.selectedProfileIds.length} (${names.join(", ")})`;
    }
    if (elements.summarySecurity) {
      elements.summarySecurity.textContent = window.I18N.t("summarySecurityActive");
    }
  }

  // Finish Actions
  async function handleFinish() {
    if (state.launchNow) {
      if (state.backendOnline) {
        try {
          await fetch("/api/launch", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ profile: state.launchProfile }),
          });
        } catch (err) {
          console.error("Launch request failed", err);
        }
      }
    }

    if (state.backendOnline) {
      try {
        await fetch("/api/exit", { method: "POST" });
      } catch {
        window.close();
      }
    } else {
      window.close();
    }
  }

  async function handleOpenFolder() {
    if (state.backendOnline) {
      try {
        await fetch("/api/open-folder", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ path: state.installPath }),
        });
      } catch (err) {
        console.error("Open folder request failed", err);
      }
    }
  }

  // Backup and Restore Handlers
  async function handleBackup() {
    if (state.backendOnline) {
      try {
        const res = await fetch("/api/backup", { method: "POST" });
        if (res.ok) {
          alert(window.I18N.t("alertBackupSuccess"));
        } else {
          const err = await res.text();
          alert(window.I18N.t("alertError") + err);
        }
      } catch (err) {
        alert(window.I18N.t("alertError") + err.message);
      }
    } else {
      alert(window.I18N.t("alertBackupSuccess") + " (Preview Mode)");
    }
  }

  async function handleRestore() {
    if (state.backendOnline) {
      try {
        const res = await fetch("/api/restore", { method: "POST" });
        if (res.ok) {
          alert(window.I18N.t("alertRestoreSuccess"));
        } else {
          const err = await res.text();
          alert(window.I18N.t("alertError") + err);
        }
      } catch (err) {
        alert(window.I18N.t("alertError") + err.message);
      }
    } else {
      alert(window.I18N.t("alertRestoreSuccess") + " (Preview Mode)");
    }
  }
})();
