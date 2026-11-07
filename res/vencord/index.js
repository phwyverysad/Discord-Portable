"use strict";

const electron = require("electron");
const path = require("path");
const fs = require("fs");
const Module = require("module");

// 1. Prevent EPIPE: broken pipe uncaught exception when stdout/stderr closes
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

// 3. Profile Identification and Windows Taskbar Isolation
const profileID = process.env.DISCORD_PORTABLE_PROFILE || "default";
const profileIcon = process.env.DISCORD_PORTABLE_ICON;
// Windows AppUserModelID must be UNIQUE per profile so Windows NEVER groups them together!
const customAumid = "Discord.PTB.Profile." + profileID;

if (electron.app) {
  try {
    electron.app.setAppUserModelId(customAumid);
    // Overwrite setAppUserModelId so Discord's internal code CANNOT revert it to com.squirrel.DiscordPTB.DiscordPTB
    electron.app.setAppUserModelId = function() {
      return customAumid;
    };
  } catch(e) {}

  // Hook window creation to ALWAYS apply the profile icon
  if (profileIcon && fs.existsSync(profileIcon)) {
    const nativeImg = electron.nativeImage ? electron.nativeImage.createFromPath(profileIcon) : profileIcon;

    electron.app.on("browser-window-created", (event, win) => {
      try {
        win.setIcon(nativeImg);
      } catch(e) {}

      win.on("ready-to-show", () => {
        try { win.setIcon(nativeImg); } catch(e) {}
      });
      win.on("show", () => {
        try { win.setIcon(nativeImg); } catch(e) {}
      });

      if (win.webContents) {
        win.webContents.on("page-favicon-updated", (ev) => {
          ev.preventDefault();
          try { win.setIcon(nativeImg); } catch(e) {}
        });
        win.webContents.on("did-finish-load", () => {
          try { win.setIcon(nativeImg); } catch(e) {}
        });
      }
    });
  }
}

// 4. Proxy BrowserWindow to inject options.icon into constructor
if (electron.BrowserWindow) {
  const OriginalBrowserWindow = electron.BrowserWindow;
  const nativeImg = (profileIcon && fs.existsSync(profileIcon) && electron.nativeImage) ?
    electron.nativeImage.createFromPath(profileIcon) : profileIcon;

  class CustomBrowserWindow extends OriginalBrowserWindow {
    constructor(opts) {
      opts = opts || {};
      if (nativeImg) {
        opts.icon = nativeImg;
      }
      super(opts);
      if (nativeImg) {
        try {
          this.setIcon(nativeImg);
        } catch(e) {}
      }
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

// 5. Delegate to Vencord / Discord Loader
require("./patcher.js");
