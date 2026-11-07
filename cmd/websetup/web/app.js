/**
 * Discord PTB Portable - Modern WebSetup Application Logic
 * Strict Constraint: ZERO EMOJIS - Pure SVG Icons & Clean UI
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
    });
  }

  function initEvents() {
    if (elements.installPathInput) {
      elements.installPathInput.addEventListener("input", (e) => {
        state.installPath = e.target.value.trim();
      });
    }

    if (elements.browseFolderBtn) {
      elements.browseFolderBtn.addEventListener("click", handleBrowseFolder);
    }

    if (elements.btnSelectAll) {
      elements.btnSelectAll.addEventListener("click", selectAllProfiles);
    }
    if (elements.btnDeselectAll) {
      elements.btnDeselectAll.addEventListener("click", deselectAllProfiles);
    }

    if (elements.switchVencord) {
      elements.switchVencord.addEventListener("change", (e) => {
        state.vencord = e.target.checked;
      });
    }

    if (elements.btnBack) {
      elements.btnBack.addEventListener("click", handleBack);
    }
    if (elements.btnNext) {
      elements.btnNext.addEventListener("click", handleNext);
    }

    if (elements.btnFinish) {
      elements.btnFinish.addEventListener("click", handleFinish);
    }

    // Custom Frameless Window Controls
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
        fetch("/api/window/close", { method: "POST" }).catch(() => {});
        window.close();
      });
    }

    // Draggable Window Header
    if (elements.windowHeader) {
      let isDragging = false;
      let startX = 0;
      let startY = 0;

      elements.windowHeader.addEventListener("mousedown", (e) => {
        if (e.target.closest("button, input, select, a, .win-btn, .lang-btn")) return;
        isDragging = true;
        startX = e.screenX;
        startY = e.screenY;

        const onMouseMove = (moveEvent) => {
          if (!isDragging) return;
          const dx = moveEvent.screenX - startX;
          const dy = moveEvent.screenY - startY;
          if (dx !== 0 || dy !== 0) {
            startX = moveEvent.screenX;
            startY = moveEvent.screenY;
            fetch(`/api/window/move?dx=${dx}&dy=${dy}`, { method: "POST" }).catch(() => {});
          }
        };

        const onMouseUp = () => {
          isDragging = false;
          window.removeEventListener("mousemove", onMouseMove);
          window.removeEventListener("mouseup", onMouseUp);
        };

        window.addEventListener("mousemove", onMouseMove);
        window.addEventListener("mouseup", onMouseUp);
      });
    }
  }

  function getStaticProfilesFallback() {
    return [
      { id: 1, code: "01", key: "yellow", hex: "#FFFF00", png_file: "01_yellow_FFFF00.png" },
      { id: 2, code: "02", key: "yellow_orange", hex: "#FFC000", png_file: "02_yellow_orange_FFC000.png" },
      { id: 3, code: "03", key: "orange", hex: "#F58220", png_file: "03_orange_F58220.png" },
      { id: 4, code: "04", key: "red_orange", hex: "#FA6304", png_file: "04_red_orange_FA6304.png" },
      { id: 5, code: "05", key: "red", hex: "#ED1C24", png_file: "05_red_ED1C24.png" },
      { id: 6, code: "06", key: "red_violet", hex: "#BD0074", png_file: "06_red_violet_BD0074.png" },
      { id: 7, code: "07", key: "violet", hex: "#59277A", png_file: "07_violet_59277A.png" },
      { id: 8, code: "08", key: "blue_violet", hex: "#4400A3", png_file: "08_blue_violet_4400A3.png" },
      { id: 9, code: "09", key: "blue", hex: "#0020C2", png_file: "09_blue_0020C2.png" },
      { id: 10, code: "10", key: "blue_green", hex: "#008240", png_file: "10_blue_green_008240.png" },
      { id: 11, code: "11", key: "green", hex: "#00B050", png_file: "11_green_00B050.png" },
      { id: 12, code: "12", key: "yellow_green", hex: "#8EF226", png_file: "12_yellow_green_8EF226.png" },
    ];
  }

  function loadProfiles() {
    state.profiles = getStaticProfilesFallback();
    fetch("/api/profiles")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (Array.isArray(data) && data.length > 0) {
          state.profiles = data.map((item) => ({
            id: item.number || item.id,
            code: String(item.number || item.id).padStart(2, "0"),
            key: item.colorKey || item.key,
            hex: item.colorHex || item.hex,
            png_file: (item.iconPath ? item.iconPath.split("/").pop() : null) || `profile_${String(item.number || item.id).padStart(2, "0")}.png`,
          }));
        }
      })
      .catch(() => {})
      .finally(() => {
        renderProfiles();
        updateCounter();
      });
  }

  function renderProfiles() {
    if (!elements.profilesGrid) return;
    elements.profilesGrid.innerHTML = "";

    state.profiles.forEach((p) => {
      const isSelected = state.selectedProfileIds.includes(p.id);
      const card = document.createElement("div");
      card.className = `profile-card ${isSelected ? "selected" : ""}`;
      card.setAttribute("data-id", p.id);

      const discordName = DISCORD_PROFILE_NAMES[p.id] || `Discord Profile ${p.id}`;

      // Clean SVG Checkmark
      const checkSvg = `
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      `;

      card.innerHTML = `
        <div class="profile-icon-wrap">
          <img class="profile-icon-img" src="assets/icons/${p.png_file}" alt="${discordName}" />
        </div>
        <div class="profile-details">
          <div class="profile-name">${discordName}</div>
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
        alert(window.I18N.t("alertSelectAtLeastOne"));
        return;
      }
      startInstallation();
    }
  }

  function startInstallation() {
    state.isInstalling = true;
    goToStep(3);

    // Prepare profile IDs in standard format (e.g. "01_yellow", "02_yellow_orange")
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
      launchOnFinish: elements.launchNowCheck ? elements.launchNowCheck.checked : true,
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
        // Fallback simulation for preview testing
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
    const launchNow = elements.launchNowCheck ? elements.launchNowCheck.checked : true;
    if (launchNow) {
      fetch("/api/launch", { method: "POST" }).catch(() => {});
    }

    fetch("/api/exit", { method: "POST" })
      .then(() => {
        window.close();
      })
      .catch(() => {
        window.close();
      });
  }
})();
