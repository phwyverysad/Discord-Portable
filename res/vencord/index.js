// Vencord Shim Loader for Discord PTB Portable
const electron = require("electron");
const path = require("path");
const fs = require("fs");

if (electron && electron.app) {
  const { app, nativeImage } = electron;

  // 1. Resolve Profile ID from environment or command line flags
  let profile = process.env.DISCORD_PORTABLE_PROFILE || "";
  if (!profile) {
    for (const arg of process.argv) {
      if (arg.startsWith("--profile=")) {
        profile = arg.substring(10).trim();
        break;
      } else if (arg.startsWith("--user-data-dir=")) {
        const parts = arg.split(/[\\/]/);
        const last = parts[parts.length - 1];
        if (last && last !== "data") profile = last.trim();
      }
    }
  }
  if (!profile) profile = "01_yellow";

  const aumid = `Discord.PTB.Profile.${profile}`;

  // Override app.setAppUserModelId to enforce our profile AUMID matching the Windows Shell shortcut
  try {
    app.setAppUserModelId(aumid);
    const origSetAppUserModelId = app.setAppUserModelId.bind(app);
    app.setAppUserModelId = function() {
      return origSetAppUserModelId(aumid);
    };
  } catch (e) {
    console.error("[ProfileShim] Failed to set AppUserModelId:", e);
  }

  // 2. Resolve Profile Icon (.ico)
  let iconPath = process.env.DISCORD_PORTABLE_ICON;
  if (!iconPath || !fs.existsSync(iconPath)) {
    const rootDir = path.resolve(__dirname, "..", "..");
    const cand = path.join(rootDir, "res", "icons", `${profile}.ico`);
    if (fs.existsSync(cand)) {
      iconPath = cand;
    } else {
      const iconsDir = path.join(rootDir, "res", "icons");
      if (fs.existsSync(iconsDir)) {
        const files = fs.readdirSync(iconsDir);
        const match = files.find(f => f.startsWith(profile) && f.endsWith(".ico"));
        if (match) iconPath = path.join(iconsDir, match);
      }
    }
  }

  // 3. Set window icon whenever any BrowserWindow is created
  if (iconPath && fs.existsSync(iconPath)) {
    try {
      const nIcon = nativeImage.createFromPath(iconPath);
      app.on("browser-window-created", (evt, win) => {
        try {
          win.setIcon(nIcon);
        } catch (_) {}
      });
    } catch (e) {
      console.error("[ProfileShim] Failed to prepare icon:", e);
    }
  }
}

require("./patcher.js");
