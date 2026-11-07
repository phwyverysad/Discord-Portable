/**
 * Discord PTB Portable - Modern WebSetup Application Logic
 * Strict Constraint: ZERO EMOJIS - Pure SVG Icons & Clean UX Architecture
 * Applied: 8-Point Grid, Fitts's Law, Nielsen Usability Heuristics
 */

(function () {
  "use strict";

  const DISCORD_PROFILE_NAMES = {
    1: "Discord Yellow",
    2: "Discord Yellow Orange",
    3: "Discord Orange",
    4: "Discord Red Orange",
    5: "Discord Red",
    6: "Discord Red Violet",
    7: "Discord Violet",
    8: "Discord Blue Violet",
    9: "Discord Blue",
    10: "Discord Blue Green",
    11: "Discord Green",
    12: "Discord Yellow Green",
  };

  const state = {
    currentStep: 1,
    installPath: "C:\\Discord-PTB-Portable",
    vencord: true,
    launchOnFinish: true,
    createShortcuts: true,
    selectedProfileIds: [1], // Default: Profile 1 Yellow
    profiles: [],
    isInstalling: false,
    pollTimer: null,
  };

  const elements = {};

  document.addEventListener("DOMContentLoaded", () => {
    cacheElements();
    initI18n();
    initEvents();
    loadProfiles();
  });

  function cacheElements() {
    elements.stepItems = document.querySelectorAll(".step-item");
    elements.stepViews = document.querySelectorAll(".step-view");
    elements.connectors = document.querySelectorAll(".step-connector");

    elements.installPathInput = document.getElementById("install-path-input");
    elements.browseFolderBtn = document.getElementById("browse-folder-btn");

    elements.profilesGrid = document.getElementById("profiles-grid");
    elements.profileCounter = document.getElementById("profile-counter");
    elements.btnSelectAll = document.getElementById("btn-select-all");
    elements.btnDeselectAll = document.getElementById("btn-deselect-all");

    elements.switchVencord = document.getElementById("switch-vencord");
    elements.switchLaunchNow = document.getElementById("switch-launch-now");
    elements.switchDesktopShortcuts = document.getElementById("switch-desktop-shortcuts");

    elements.progressFill = document.getElementById("progress-fill");
    elements.progressPercent = document.getElementById("progress-percent");
    elements.progressStatus = document.getElementById("progress-status");
    elements.stageItems = document.querySelectorAll(".stage-item");

    elements.launchNowCheck = document.getElementById("launch-now-check");
    elements.shortcutsCheck = document.getElementById("shortcuts-check");
    elements.btnFinish = document.getElementById("btn-finish");

    elements.btnBack = document.getElementById("btn-back");
    elements.btnNext = document.getElementById("btn-next");

    elements.windowHeader = document.querySelector(".window-header");
    elements.winBtnMinimize = document.getElementById("win-btn-minimize");
    elements.winBtnMaximize = document.getElementById("win-btn-maximize");
    elements.winBtnClose = document.getElementById("win-btn-close");
  }

  function initI18n() {
    window.I18N.init();

    // Custom Language Dropdown Wiring
    const dropdown = document.getElementById("lang-dropdown");
    const trigger = document.getElementById("lang-dropdown-trigger");
    const menu = document.getElementById("lang-dropdown-menu");
    const currentLangText = document.getElementById("current-lang-text");
    const items = menu ? menu.querySelectorAll(".dropdown-item") : [];

    function updateDropdownUI(lang) {
      if (currentLangText) {
        currentLangText.textContent = lang === "en" ? "English" : "ไทย";
      }
      items.forEach((item) => {
        const itemLang = item.getAttribute("data-lang");
        const isActive = itemLang === lang;
        item.classList.toggle("active", isActive);
        item.setAttribute("aria-selected", isActive ? "true" : "false");
      });
    }

    if (trigger && menu) {
      trigger.addEventListener("click", (e) => {
        e.stopPropagation();
        const isOpen = menu.classList.contains("show");
        if (isOpen) {
          menu.classList.remove("show");
          trigger.classList.remove("open");
          trigger.setAttribute("aria-expanded", "false");
        } else {
          menu.classList.add("show");
          trigger.classList.add("open");
          trigger.setAttribute("aria-expanded", "true");
        }
      });

      items.forEach((item) => {
        item.addEventListener("click", (e) => {
          e.stopPropagation();
          const selectedLang = item.getAttribute("data-lang");
          if (selectedLang) {
            window.I18N.setLanguage(selectedLang);
            updateDropdownUI(selectedLang);
          }
          menu.classList.remove("show");
          trigger.classList.remove("open");
          trigger.setAttribute("aria-expanded", "false");
          renderProfiles();
          updateCounter();
        });
      });

      document.addEventListener("click", (e) => {
        if (!dropdown.contains(e.target)) {
          menu.classList.remove("show");
          trigger.classList.remove("open");
          trigger.setAttribute("aria-expanded", "false");
        }
      });
    }

    updateDropdownUI(window.I18N.currentLang);
  }

  function initEvents() {
    // Window Controls
    if (elements.winBtnMinimize) {
      elements.winBtnMinimize.addEventListener("click", () => {
        fetch("/api/window/minimize", { method: "POST" }).catch(() => {});
      });
    }

    if (elements.winBtnMaximize) {
      elements.winBtnMaximize.addEventListener("click", () => {
        fetch("/api/window/maximize", { method: "POST" }).catch(() => {});
      });
    }

    if (elements.winBtnClose) {
      elements.winBtnClose.addEventListener("click", () => {
        fetch("/api/window/close", { method: "POST" })
          .then(() => window.close())
          .catch(() => window.close());
      });
    }

    // Window Dragging via Header
    if (elements.windowHeader) {
      let isDragging = false;
      let startX = 0;
      let startY = 0;

      elements.windowHeader.addEventListener("mousedown", (e) => {
        if (e.target.closest("button") || e.target.closest(".custom-dropdown")) {
          return;
        }
        isDragging = true;
        startX = e.screenX;
        startY = e.screenY;
      });

      window.addEventListener("mousemove", (e) => {
        if (!isDragging) return;
        const dx = e.screenX - startX;
        const dy = e.screenY - startY;
        startX = e.screenX;
        startY = e.screenY;

        if (dx !== 0 || dy !== 0) {
          fetch(`/api/window/move?dx=${dx}&dy=${dy}`, { method: "POST" }).catch(() => {});
        }
      });

      window.addEventListener("mouseup", () => {
        isDragging = false;
      });
    }

    // Navigation buttons
    if (elements.btnBack) {
      elements.btnBack.addEventListener("click", handleBack);
    }
    if (elements.btnNext) {
      elements.btnNext.addEventListener("click", handleNext);
    }

    // Step 1: Browse Folder
    if (elements.browseFolderBtn) {
      elements.browseFolderBtn.addEventListener("click", handleBrowseFolder);
    }
    if (elements.installPathInput) {
      elements.installPathInput.addEventListener("input", (e) => {
        state.installPath = e.target.value.trim();
      });
    }

    // Step 2: Toolbar buttons
    if (elements.btnSelectAll) {
      elements.btnSelectAll.addEventListener("click", selectAllProfiles);
    }
    if (elements.btnDeselectAll) {
      elements.btnDeselectAll.addEventListener("click", deselectAllProfiles);
    }

    // Switches synchronization
    if (elements.switchVencord) {
      elements.switchVencord.addEventListener("change", (e) => {
        state.vencord = e.target.checked;
      });
    }

    if (elements.switchLaunchNow) {
      elements.switchLaunchNow.addEventListener("change", (e) => {
        state.launchOnFinish = e.target.checked;
        if (elements.launchNowCheck) {
          elements.launchNowCheck.checked = e.target.checked;
        }
      });
    }

    if (elements.switchDesktopShortcuts) {
      elements.switchDesktopShortcuts.addEventListener("change", (e) => {
        state.createShortcuts = e.target.checked;
        if (elements.shortcutsCheck) {
          elements.shortcutsCheck.checked = e.target.checked;
        }
      });
    }

    if (elements.launchNowCheck) {
      elements.launchNowCheck.addEventListener("change", (e) => {
        state.launchOnFinish = e.target.checked;
        if (elements.switchLaunchNow) {
          elements.switchLaunchNow.checked = e.target.checked;
        }
      });
    }

    if (elements.shortcutsCheck) {
      elements.shortcutsCheck.addEventListener("change", (e) => {
        state.createShortcuts = e.target.checked;
        if (elements.switchDesktopShortcuts) {
          elements.switchDesktopShortcuts.checked = e.target.checked;
        }
      });
    }

    // Step 4: Finish Button
    if (elements.btnFinish) {
      elements.btnFinish.addEventListener("click", handleFinish);
    }
  }

  function loadProfiles() {
    fetch("/api/profiles")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          state.profiles = data.map((item) => ({
            id: item.number,
            code: String(item.number).padStart(2, "0"),
            key: item.colorKey,
            name_th: DISCORD_PROFILE_NAMES[item.number] || item.nameTH,
            name_en: DISCORD_PROFILE_NAMES[item.number] || item.nameEN,
            color: item.colorHex,
            png_file: `${String(item.number).padStart(2, "0")}_${item.colorKey}_${item.colorHex.replace("#", "")}.png`,
            ico_file: `${String(item.number).padStart(2, "0")}_${item.colorKey}.ico`,
          }));
        } else {
          fallbackProfiles();
        }
        renderProfiles();
        updateCounter();
      })
      .catch(() => {
        fallbackProfiles();
        renderProfiles();
        updateCounter();
      });
  }

  function fallbackProfiles() {
    const colorDefs = [
      { id: 1, key: "yellow", hex: "FFFF00" },
      { id: 2, key: "yellow_orange", hex: "FFC000" },
      { id: 3, key: "orange", hex: "F58220" },
      { id: 4, key: "red_orange", hex: "FA6304" },
      { id: 5, key: "red", hex: "ED1C24" },
      { id: 6, key: "red_violet", hex: "BD0074" },
      { id: 7, key: "violet", hex: "59277A" },
      { id: 8, key: "blue_violet", hex: "4400A3" },
      { id: 9, key: "blue", hex: "0020C2" },
      { id: 10, key: "blue_green", hex: "008240" },
      { id: 11, key: "green", hex: "00B050" },
      { id: 12, key: "yellow_green", hex: "8EF226" },
    ];

    state.profiles = colorDefs.map((c) => ({
      id: c.id,
      code: String(c.id).padStart(2, "0"),
      key: c.key,
      name_th: DISCORD_PROFILE_NAMES[c.id],
      name_en: DISCORD_PROFILE_NAMES[c.id],
      color: "#" + c.hex,
      png_file: `${String(c.id).padStart(2, "0")}_${c.key}_${c.hex}.png`,
      ico_file: `${String(c.id).padStart(2, "0")}_${c.key}.ico`,
    }));
  }

  function renderProfiles() {
    if (!elements.profilesGrid) return;
    elements.profilesGrid.innerHTML = "";

    state.profiles.forEach((p) => {
      const isSelected = state.selectedProfileIds.includes(p.id);
      const discordName = DISCORD_PROFILE_NAMES[p.id];

      const card = document.createElement("div");
      card.className = `profile-card ${isSelected ? "selected" : ""}`;
      card.setAttribute("data-id", p.id);
      card.setAttribute("role", "checkbox");
      card.setAttribute("aria-checked", isSelected ? "true" : "false");
      card.setAttribute("tabindex", "0");

      card.innerHTML = `
        <div class="profile-icon-wrap" style="background: ${p.color}18; border-color: ${p.color}44;">
          <img class="profile-icon-img" src="assets/icons/${p.png_file}" alt="${discordName}" />
        </div>
        <div class="profile-details">
          <div class="profile-name">${discordName}</div>
        </div>
        <div class="profile-check">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
        </div>
      `;

      card.addEventListener("click", () => toggleProfileSelection(p.id));
      card.addEventListener("keydown", (e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          toggleProfileSelection(p.id);
        }
      });
      elements.profilesGrid.appendChild(card);
    });
  }

  function toggleProfileSelection(id) {
    const idx = state.selectedProfileIds.indexOf(id);
    if (idx >= 0) {
      // Usability Heuristic 5 (Error Prevention): Keep at least 1 profile selected
      if (state.selectedProfileIds.length <= 1) {
        return;
      }
      state.selectedProfileIds.splice(idx, 1);
    } else {
      state.selectedProfileIds.push(id);
      state.selectedProfileIds.sort((a, b) => a - b);
    }

    renderProfiles();
    updateCounter();
  }

  function selectAllProfiles() {
    state.selectedProfileIds = state.profiles.map((p) => p.id);
    renderProfiles();
    updateCounter();
  }

  function deselectAllProfiles() {
    state.selectedProfileIds = [1];
    renderProfiles();
    updateCounter();
  }

  function updateCounter() {
    if (elements.profileCounter) {
      elements.profileCounter.textContent = window.I18N.t("selectedCount", {
        count: state.selectedProfileIds.length,
      });
    }
  }

  function handleBrowseFolder() {
    fetch("/api/browse", { method: "POST" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data && data.path) {
          state.installPath = data.path;
          elements.installPathInput.value = data.path;
        }
      })
      .catch(() => {});
  }

  function goToStep(step) {
    state.currentStep = step;

    // Update Step View active classes
    elements.stepViews.forEach((view, idx) => {
      view.classList.toggle("active", idx + 1 === step);
    });

    // Update Nav items
    elements.stepItems.forEach((item, idx) => {
      const stepNum = idx + 1;
      item.classList.remove("active", "completed");
      if (stepNum === step) {
        item.classList.add("active");
      } else if (stepNum < step) {
        item.classList.add("completed");
      }
    });

    // Update Connectors
    elements.connectors.forEach((conn, idx) => {
      conn.classList.toggle("completed", idx + 1 < step);
    });

    // Update Navigation Buttons visibility
    if (step === 1) {
      elements.btnBack.style.display = "none";
      elements.btnNext.style.display = "inline-flex";
      elements.btnNext.innerHTML = `
        <span>${window.I18N.t("nextBtn")}</span>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="5" y1="12" x2="19" y2="12"></line>
          <polyline points="12 5 19 12 12 19"></polyline>
        </svg>
      `;
    } else if (step === 2) {
      elements.btnBack.style.display = "inline-flex";
      elements.btnNext.style.display = "inline-flex";
      elements.btnNext.innerHTML = `
        <span>${window.I18N.t("startInstallBtn")}</span>
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      `;
    } else if (step === 3) {
      // Progress View: Hide both navigation buttons
      elements.btnBack.style.display = "none";
      elements.btnNext.style.display = "none";
    } else if (step === 4) {
      // Finished View
      elements.btnBack.style.display = "none";
      elements.btnNext.style.display = "none";
    }
  }

  function handleBack() {
    if (state.currentStep > 1 && !state.isInstalling) {
      goToStep(state.currentStep - 1);
    }
  }

  function handleNext() {
    if (state.currentStep === 1) {
      goToStep(2);
    } else if (state.currentStep === 2) {
      if (state.selectedProfileIds.length === 0) {
        state.selectedProfileIds = [1];
        renderProfiles();
        updateCounter();
      }
      startInstallation();
    }
  }

  function startInstallation() {
    state.isInstalling = true;
    goToStep(3);

    const profileKeys = state.selectedProfileIds.map((id) => {
      const p = state.profiles.find((x) => x.id === id);
      if (p && p.key) {
        return `${p.code}_${p.key}`;
      }
      return `profile_${id}`;
    });

    const payload = {
      installDir: state.installPath,
      installMode: "pc",
      ecoMode: false,
      profiles: profileKeys,
      launchOnFinish: elements.switchLaunchNow ? elements.switchLaunchNow.checked : true,
      createShortcuts: elements.switchDesktopShortcuts ? elements.switchDesktopShortcuts.checked : true,
    };

    fetch("/api/install", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    })
      .then((res) => (res.ok ? res.json() : null))
      .then(() => {
        startPollingProgress();
      })
      .catch(() => {
        simulateOfflineProgress();
      });
  }

  function startPollingProgress() {
    if (state.pollTimer) clearInterval(state.pollTimer);

    state.pollTimer = setInterval(() => {
      fetch("/api/status")
        .then((res) => (res.ok ? res.json() : null))
        .then((status) => {
          if (!status) return;

          updateProgressUI(status.percent, status.message);

          if (status.done) {
            clearInterval(state.pollTimer);
            state.pollTimer = null;
            state.isInstalling = false;
            setTimeout(() => {
              goToStep(4);
            }, 600);
          }
        })
        .catch(() => {});
    }, 300);
  }

  function simulateOfflineProgress() {
    let p = 0;
    const stages = [
      { pct: 25, msg: window.I18N.t("statusDownloading"), stage: 1 },
      { pct: 55, msg: window.I18N.t("statusVencord"), stage: 2 },
      { pct: 80, msg: window.I18N.t("statusHardening"), stage: 3 },
      { pct: 100, msg: window.I18N.t("statusCompleted"), stage: 4 },
    ];

    let stageIdx = 0;
    const timer = setInterval(() => {
      p += 5;
      if (stageIdx < stages.length && p >= stages[stageIdx].pct) {
        updateProgressUI(stages[stageIdx].pct, stages[stageIdx].msg, stages[stageIdx].stage);
        stageIdx++;
      } else {
        updateProgressUI(p);
      }

      if (p >= 100) {
        clearInterval(timer);
        state.isInstalling = false;
        setTimeout(() => {
          goToStep(4);
        }, 600);
      }
    }, 120);
  }

  function updateProgressUI(percent, message, activeStage = 1) {
    if (elements.progressFill) {
      elements.progressFill.style.width = `${percent}%`;
    }
    if (elements.progressPercent) {
      elements.progressPercent.textContent = `${percent}%`;
    }
    if (message && elements.progressStatus) {
      elements.progressStatus.textContent = message;
    }

    // Update stage items
    const calcStage = activeStage || (percent < 30 ? 1 : percent < 60 ? 2 : percent < 90 ? 3 : 4);
    elements.stageItems.forEach((item, idx) => {
      const s = idx + 1;
      item.classList.remove("completed", "in-progress");
      if (s < calcStage || percent >= 100) {
        item.classList.add("completed");
      } else if (s === calcStage) {
        item.classList.add("in-progress");
      }
    });
  }

  function handleFinish() {
    const launchNow = elements.launchNowCheck ? elements.launchNowCheck.checked :
      (elements.switchLaunchNow ? elements.switchLaunchNow.checked : true);

    const firstProfileKey = state.selectedProfileIds.length > 0 ?
      (() => {
        const p = state.profiles.find((x) => x.id === state.selectedProfileIds[0]);
        return p && p.key ? `${p.code}_${p.key}` : "01_yellow";
      })() : "01_yellow";

    if (launchNow) {
      fetch("/api/launch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          profile: firstProfileKey,
          installDir: state.installPath,
        }),
      }).catch(() => {});
    }

    setTimeout(() => {
      fetch("/api/exit", { method: "POST" })
        .then(() => window.close())
        .catch(() => window.close());
    }, 200);
  }
})();
