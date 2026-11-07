"use strict";

const electron = require("electron");
const path = require("path");
const fs = require("fs");
const Module = require("module");

// 1. Prevent EPIPE: broken pipe uncaught exception when stdout/stderr closes on exit
function suppressEpipe(stream) {
  if (!stream) return;
  const origWrite = stream.write;
  stream.write = function(chunk, enc, cb) {
    try {
      return origWrite.call(stream, chunk, enc, (err) => {
        if (err && (err.code === "EPIPE" || (err.message && err.message.includes("EPIPE")))) {
          if (typeof cb === "function") cb();
          return;
        }
        if (typeof cb === "function") cb(err);
      });
    } catch (err) {
      if (err && (err.code === "EPIPE" || (err.message && err.message.includes("EPIPE")))) {
        if (typeof cb === "function") cb();
        return true;
      }
      throw err;
    }
  };
  stream.on("error", (err) => {
    if (err && (err.code === "EPIPE" || (err.message && err.message.includes("EPIPE")))) return;
  });
}
suppressEpipe(process.stdout);
suppressEpipe(process.stderr);

for (const m of ["log", "info", "warn", "error", "debug"]) {
  if (console[m]) {
    const orig = console[m];
    console[m] = function(...args) {
      try {
        orig.apply(console, args);
      } catch (err) {
        if (err && (err.code === "EPIPE" || (err.message && err.message.includes("EPIPE")))) return;
      }
    };
  }
}

process.on("uncaughtException", (err) => {
  if (err && (err.code === "EPIPE" || (err.message && err.message.includes("EPIPE")))) return;
  console.error("[Discord-Portable] Uncaught exception:", err);
});

// 2. Ensure Discord native updater committed modules are resolvable
const exeDir = path.dirname(process.execPath);
const modulesDir = path.join(exeDir, "modules");
if (fs.existsSync(modulesDir)) {
  try {
    const entries = fs.readdirSync(modulesDir);
    for (const entry of entries) {
      const full = path.join(modulesDir, entry);
      if (fs.statSync(full).isDirectory()) {
        Module.globalPaths.push(full);
      }
    }
  } catch(e) {}
}

// 3. Multi-Strategy Profile & Icon Resolution
function detectProfileAndIcon() {
  let profile = process.env.DISCORD_PORTABLE_PROFILE;
  let iconPath = process.env.DISCORD_PORTABLE_ICON;

  if (!profile) {
    for (const arg of process.argv) {
      if (arg.startsWith("--profile=")) {
        profile = arg.slice("--profile=".length).trim();
        break;
      }
      if (arg.startsWith("--user-data-dir=")) {
        const udir = arg.slice("--user-data-dir=".length).replace(/"/g, "").trim();
        profile = path.basename(udir);
        break;
      }
    }
  }
  if (!profile) profile = "01_yellow";

  // If icon path not in env or doesn't exist, search up the folder tree
  if (!iconPath || !fs.existsSync(iconPath)) {
    const candidateDirs = [
      path.join(__dirname, "../../../res/icons"),       // relative from app/app-1.0.x/resources/app
      path.join(__dirname, "../../../../res/icons"),
      path.join(path.dirname(process.execPath), "../../res/icons"),
      path.join(path.dirname(process.execPath), "../res/icons"),
      path.join(process.cwd(), "res/icons"),
    ];

    for (const cDir of candidateDirs) {
      if (!fs.existsSync(cDir)) continue;
      // Direct file match
      const directIco = path.join(cDir, `${profile}.ico`);
      if (fs.existsSync(directIco)) {
        iconPath = directIco;
        break;
      }
      // Match by prefix (e.g. "05" or "05_red")
      try {
        const list = fs.readdirSync(cDir);
        for (const f of list) {
          if (f.toLowerCase().endsWith(".ico")) {
            if (f.startsWith(profile + "_") || f.startsWith(profile + ".") || f.startsWith(profile.split("_")[0] + "_")) {
              iconPath = path.join(cDir, f);
              break;
            }
          }
        }
        if (iconPath) break;
      } catch(e) {}
    }
  }

  return { profile, iconPath };
}

const { profile, iconPath } = detectProfileAndIcon();
const customAumid = "Discord.PTB.Profile." + profile;

// Store on globalThis for downstream hooks / patcher
globalThis._discordPortableProfile = profile;
globalThis._discordPortableIconPath = iconPath;
globalThis._discordPortableAumid = customAumid;

// 4. Enforce Unique AUMID in Windows Taskbar
if (electron.app) {
  try {
    electron.app.setAppUserModelId(customAumid);
    Object.defineProperty(electron.app, "setAppUserModelId", {
      value: function() { return customAumid; },
      writable: false,
      configurable: true
    });
  } catch(e) {}
}

// 5. Native Image Preparation
let nativeImg = null;
if (iconPath && fs.existsSync(iconPath) && electron.nativeImage) {
  try {
    nativeImg = electron.nativeImage.createFromPath(iconPath);
    globalThis._discordPortableNativeImg = nativeImg;
  } catch(e) {}
}

function enforceWindowProfile(win) {
  if (!win || win.isDestroyed()) return;
  if (nativeImg) {
    try { win.setIcon(nativeImg); } catch(e) {}
  }
  try {
    if (typeof win.setAppDetails === "function") {
      win.setAppDetails({
        appId: customAumid,
        appIconPath: iconPath || process.execPath,
        appIconIndex: 0,
        relaunchDisplayName: "Discord " + profile,
      });
    }
  } catch(e) {}
}

// 6. Hook browser-window-created to enforce icon across all states
if (electron.app) {
  electron.app.on("browser-window-created", (event, win) => {
    enforceWindowProfile(win);

    win.on("ready-to-show", () => enforceWindowProfile(win));
    win.on("show", () => enforceWindowProfile(win));
    win.on("focus", () => enforceWindowProfile(win));

    if (win.webContents) {
      // Intercept page favicon updates so notification unread badges DO NOT revert to blurple!
      win.webContents.on("page-favicon-updated", (ev) => {
        ev.preventDefault();
        enforceWindowProfile(win);
      });
      win.webContents.on("did-finish-load", () => enforceWindowProfile(win));
      win.webContents.on("dom-ready", () => enforceWindowProfile(win));
    }
  });
}

// 7. Proxy BrowserWindow constructor
if (electron.BrowserWindow) {
  const OriginalBrowserWindow = electron.BrowserWindow;

  class CustomBrowserWindow extends OriginalBrowserWindow {
    constructor(opts) {
      opts = opts || {};
      if (nativeImg) {
        opts.icon = nativeImg;
      }
      super(opts);
      enforceWindowProfile(this);
    }
  }
  Object.assign(CustomBrowserWindow, OriginalBrowserWindow);
  Object.defineProperty(CustomBrowserWindow, "name", { value: "BrowserWindow", configurable: true });

  electron.BrowserWindow = CustomBrowserWindow;
  const electronPath = require.resolve("electron");
  if (require.cache[electronPath]) {
    require.cache[electronPath].exports = {
      ...electron,
      BrowserWindow: CustomBrowserWindow,
    };
  }
}

// 8. Handoff to Vencord or Vanilla Discord
const isVanilla = process.argv.includes("--vanilla");
const patcherPath = path.join(__dirname, "patcher.js");

if (!isVanilla && fs.existsSync(patcherPath)) {
  require("./patcher.js");
} else {
  // Pure Discord mode (Vanilla)
  console.log("[Discord-Portable] Running original Discord (Profile: " + profile + ")");
  const asarPath = path.join(__dirname, "..", "_app.asar");
  const targetAsar = fs.existsSync(asarPath) ? asarPath : path.join(__dirname, "..", "app.asar");
  const pkg = require(path.join(targetAsar, "package.json"));
  require.main.filename = path.join(targetAsar, pkg.main);
  electron.app.setAppPath(targetAsar);
  require(require.main.filename);
}
