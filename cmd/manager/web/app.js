"use strict";

let managerState = {
  action: "add", // "add" or "uninstall"
  installedProfiles: [],
  allProfiles: [],
  selected: new Set()
};

async function init() {
  setupWindowControls();

  try {
    const res = await fetch("/api/state");
    const data = await res.json();
    managerState.action = data.action || "add";
    managerState.installedProfiles = data.installedProfiles || [];
    managerState.allProfiles = data.allProfiles || [];

    renderUI();
  } catch (err) {
    console.error("Failed to load manager state:", err);
  }
}

function renderUI() {
  const isUninstall = managerState.action === "uninstall";
  const headerTitle = document.getElementById("header-title");
  const headerDesc = document.getElementById("header-desc");
  const btnAction = document.getElementById("btn-action");
  const btnActionText = document.getElementById("btn-action-text");
  const windowTitle = document.getElementById("window-title-label");

  if (isUninstall) {
    windowTitle.textContent = "Discord PTB - ถอนการติดตั้งโปรไฟล์";
    headerTitle.textContent = "ถอนการติดตั้งโปรไฟล์";
    headerDesc.textContent = "เลือกโปรไฟล์ที่ต้องการลบออกจากระบบ";
    btnAction.className = "btn btn-danger";
    btnActionText.textContent = "ถอนการติดตั้ง";

    // In uninstall mode, default select all currently installed profiles
    managerState.selected.clear();
    for (const id of managerState.installedProfiles) {
      managerState.selected.add(id);
    }
  } else {
    windowTitle.textContent = "Discord PTB - ติดตั้งโปรไฟล์เพิ่มเติม";
    headerTitle.textContent = "เลือกโปรไฟล์สี";
    headerDesc.textContent = "เลือกสีประจำบัญชีเพื่อแยกไอคอนและข้อมูล";
    btnAction.className = "btn btn-primary";
    btnActionText.textContent = "ติดตั้ง";

    // In add mode, select the first uninstalled profile if available
    managerState.selected.clear();
    for (const p of managerState.allProfiles) {
      if (!managerState.installedProfiles.includes(p.id)) {
        managerState.selected.add(p.id);
        break;
      }
    }
  }

  renderGrid();
  updateBadge();
}

function renderGrid() {
  const grid = document.getElementById("profiles-grid");
  grid.innerHTML = "";
  const isUninstall = managerState.action === "uninstall";

  managerState.allProfiles.forEach(p => {
    const isInstalled = managerState.installedProfiles.includes(p.id);
    const isSelected = managerState.selected.has(p.id);

    const card = document.createElement("div");
    card.className = "profile-card";
    if (isSelected) {
      card.classList.add("selected");
      if (isUninstall) card.classList.add("mode-uninstall");
    }

    card.innerHTML = `
      <div class="profile-card-left">
        <div class="profile-icon-circle">
          <img src="${p.iconPath}" alt="${p.nameTH}" />
        </div>
        <div class="profile-name">
          ${p.nameTH}
          ${!isUninstall && isInstalled ? '<span class="profile-status-tag">ติดตั้งแล้ว</span>' : ''}
        </div>
      </div>
      <div class="profile-check">
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
      </div>
    `;

    card.addEventListener("click", () => {
      toggleProfile(p.id);
    });

    grid.appendChild(card);
  });
}

function toggleProfile(id) {
  if (managerState.selected.has(id)) {
    managerState.selected.delete(id);
  } else {
    managerState.selected.add(id);
  }
  renderGrid();
  updateBadge();
}

function updateBadge() {
  const count = managerState.selected.size;
  const badge = document.getElementById("selected-badge");
  badge.textContent = `เลือกแล้ว ${count} บัญชี`;

  const btnAction = document.getElementById("btn-action");
  btnAction.disabled = count === 0;

  const isUninstall = managerState.action === "uninstall";
  const notice = document.getElementById("uninstall-notice");
  const isAll = (count === managerState.allProfiles.length) || 
                (managerState.installedProfiles.length > 0 && managerState.installedProfiles.every(id => managerState.selected.has(id)));

  if (isUninstall && isAll) {
    notice.style.display = "flex";
  } else {
    notice.style.display = "none";
  }
}

// Select All / Deselect All
document.getElementById("btn-select-all").addEventListener("click", () => {
  managerState.allProfiles.forEach(p => managerState.selected.add(p.id));
  renderGrid();
  updateBadge();
});

document.getElementById("btn-deselect-all").addEventListener("click", () => {
  managerState.selected.clear();
  renderGrid();
  updateBadge();
});

// Action Execution (Install or Uninstall)
document.getElementById("btn-action").addEventListener("click", async () => {
  const btn = document.getElementById("btn-action");
  const btnText = document.getElementById("btn-action-text");
  const isUninstall = managerState.action === "uninstall";

  btn.disabled = true;
  btnText.textContent = isUninstall ? "กำลังถอนการติดตั้ง..." : "กำลังติดตั้ง...";

  const count = managerState.selected.size;
  const isAll = (count === managerState.allProfiles.length) || 
                (managerState.installedProfiles.length > 0 && managerState.installedProfiles.every(id => managerState.selected.has(id)));

  const payload = {
    action: managerState.action,
    profiles: Array.from(managerState.selected),
    all: isUninstall ? isAll : false
  };

  try {
    const res = await fetch("/api/execute", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload)
    });
    const result = await res.json();
    if (result.success) {
      btnText.textContent = isUninstall ? "ถอนการติดตั้งเรียบร้อย" : "ติดตั้งสำเร็จ";
      setTimeout(() => {
        fetch("/api/exit", { method: "POST" });
      }, 500);
    } else {
      alert("เกิดข้อผิดพลาด: " + (result.error || "ไม่สามารถดำเนินการได้"));
      btn.disabled = false;
      btnText.textContent = isUninstall ? "ถอนการติดตั้ง" : "ติดตั้ง";
    }
  } catch (err) {
    alert("เกิดข้อผิดพลาดในการเชื่อมต่อ: " + err.message);
    btn.disabled = false;
    btnText.textContent = isUninstall ? "ถอนการติดตั้ง" : "ติดตั้ง";
  }
});

// Cancel / Close
document.getElementById("btn-cancel").addEventListener("click", () => {
  fetch("/api/exit", { method: "POST" });
});

document.getElementById("win-btn-close").addEventListener("click", () => {
  fetch("/api/exit", { method: "POST" });
});

document.getElementById("win-btn-minimize").addEventListener("click", () => {
  fetch("/api/window/minimize", { method: "POST" });
});

function setupWindowControls() {
  const dragHandle = document.getElementById("drag-region");
  let isDragging = false;
  let startX = 0;
  let startY = 0;

  dragHandle.addEventListener("mousedown", (e) => {
    if (e.target.closest(".window-controls-cluster") || e.target.closest("button")) return;
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
      fetch(`/api/window/move?dx=${dx}&dy=${dy}`);
    }
  });

  window.addEventListener("mouseup", () => {
    isDragging = false;
  });
}

document.addEventListener("DOMContentLoaded", init);
