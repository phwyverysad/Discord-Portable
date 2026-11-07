"""
Real Machine Window Icon Automated Verification Test
Verifies that Discord PTB Portable windows on the host machine carry
the exact profile-specific colored icons (not the default blurple icon).
"""

import os
import sys
import time
import subprocess
import ctypes
from ctypes import wintypes
from PIL import Image
import numpy as np

user32 = ctypes.windll.user32
gdi32 = ctypes.windll.gdi32

WM_GETICON = 0x007F
ICON_SMALL = 0
ICON_BIG = 1

gdi32.GetDIBits.argtypes = [wintypes.HDC, wintypes.HBITMAP, wintypes.UINT, wintypes.UINT, ctypes.c_void_p, ctypes.c_void_p, wintypes.UINT]
gdi32.GetDIBits.restype = ctypes.c_int
gdi32.DeleteObject.argtypes = [wintypes.HGDIOBJ]
gdi32.DeleteDC.argtypes = [wintypes.HDC]
user32.ReleaseDC.argtypes = [wintypes.HWND, wintypes.HDC]
user32.GetIconInfo.argtypes = [wintypes.HICON, ctypes.c_void_p]
user32.GetIconInfo.restype = wintypes.BOOL

class ICONINFO(ctypes.Structure):
    _fields_ = [
        ('fIcon', wintypes.BOOL),
        ('xHotspot', wintypes.DWORD),
        ('yHotspot', wintypes.DWORD),
        ('hbmMask', wintypes.HBITMAP),
        ('hbmColor', wintypes.HBITMAP)
    ]

class BITMAPINFOHEADER(ctypes.Structure):
    _pack_ = 1
    _fields_ = [
        ('biSize', wintypes.DWORD),
        ('biWidth', wintypes.LONG),
        ('biHeight', wintypes.LONG),
        ('biPlanes', wintypes.WORD),
        ('biBitCount', wintypes.WORD),
        ('biCompression', wintypes.DWORD),
        ('biSizeImage', wintypes.DWORD),
        ('biXPelsPerMeter', wintypes.LONG),
        ('biYPelsPerMeter', wintypes.LONG),
        ('biClrUsed', wintypes.DWORD),
        ('biClrImportant', wintypes.DWORD)
    ]

def hicon_to_image(hicon):
    if not hicon:
        return None
    ii = ICONINFO()
    if not user32.GetIconInfo(hicon, ctypes.byref(ii)):
        return None
    
    hdc = user32.GetDC(0)
    hdc_mem = gdi32.CreateCompatibleDC(hdc)
    
    bmi = BITMAPINFOHEADER()
    bmi.biSize = ctypes.sizeof(BITMAPINFOHEADER)
    gdi32.GetDIBits(hdc_mem, ii.hbmColor, 0, 0, None, ctypes.byref(bmi), 0)
    
    w = bmi.biWidth
    h = abs(bmi.biHeight)
    bmi.biHeight = -h
    bmi.biBitCount = 32
    bmi.biCompression = 0
    
    buf = ctypes.create_string_buffer(w * h * 4)
    gdi32.GetDIBits(hdc_mem, ii.hbmColor, 0, h, buf, ctypes.byref(bmi), 0)
    
    arr = np.frombuffer(buf, dtype=np.uint8).reshape((h, w, 4))
    rgba = np.zeros((h, w, 4), dtype=np.uint8)
    rgba[:, :, 0] = arr[:, :, 2] # Red
    rgba[:, :, 1] = arr[:, :, 1] # Green
    rgba[:, :, 2] = arr[:, :, 0] # Blue
    rgba[:, :, 3] = arr[:, :, 3] # Alpha
    
    gdi32.DeleteObject(ii.hbmMask)
    gdi32.DeleteObject(ii.hbmColor)
    gdi32.DeleteDC(hdc_mem)
    user32.ReleaseDC(0, hdc)
    
    return Image.fromarray(rgba, 'RGBA')

def find_discord_main_windows():
    found = []
    def enum_proc(hwnd, lParam):
        if not user32.IsWindowVisible(hwnd):
            return True
        length = user32.GetWindowTextLengthW(hwnd)
        if length > 0:
            buff = ctypes.create_unicode_buffer(length + 1)
            user32.GetWindowTextW(hwnd, buff, length + 1)
            title = buff.value
            
            c_len = 256
            class_buff = ctypes.create_unicode_buffer(c_len)
            user32.GetClassNameW(hwnd, class_buff, c_len)
            class_name = class_buff.value
            
            pid = wintypes.DWORD()
            user32.GetWindowThreadProcessId(hwnd, ctypes.byref(pid))
            
            if "discord" in title.lower() or "discord" in class_name.lower():
                rect = wintypes.RECT()
                user32.GetWindowRect(hwnd, ctypes.byref(rect))
                w = rect.right - rect.left
                h = rect.bottom - rect.top
                if w > 400 and h > 400 and "updater" not in title.lower():
                    found.append((hwnd, pid.value, title, class_name, w, h))
        return True

    WNDENUMPROC = ctypes.WINFUNCTYPE(ctypes.c_bool, wintypes.HWND, wintypes.LPARAM)
    user32.EnumWindows(WNDENUMPROC(enum_proc), 0)
    return found

def verify_profile_window(profile_id, expected_color):
    print(f"\n[TEST] Verifying Real Window for Profile: {profile_id} (Expected: {expected_color})")
    
    subprocess.call(["taskkill", "/F", "/IM", "DiscordPTB.exe"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    subprocess.call(["taskkill", "/F", "/IM", "DiscordPTB-Portable.exe"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    time.sleep(1)
    
    exe_path = os.path.abspath(r"C:\Discord-PTB-Portable\DiscordPTB-Portable.exe")
    if not os.path.exists(exe_path):
        exe_path = os.path.abspath("DiscordPTB-Portable.exe")
        
    proc = subprocess.Popen([exe_path, f"--profile={profile_id}"])
    
    main_hwnd = None
    start = time.time()
    while time.time() - start < 45:
        time.sleep(1.5)
        wins = find_discord_main_windows()
        if wins:
            main_hwnd = wins[0][0]
            print(f"  Live HWND: 0x{main_hwnd:X} | Title: '{wins[0][2]}'")
            break
            
    if not main_hwnd:
        print("  FAIL: Window not found within 45 seconds")
        subprocess.call(["taskkill", "/F", "/T", "/PID", str(proc.pid)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return False
        
    time.sleep(3)
    
    hicon_big = user32.SendMessageW(main_hwnd, WM_GETICON, ICON_BIG, 0)
    hicon_small = user32.SendMessageW(main_hwnd, WM_GETICON, ICON_SMALL, 0)
    
    print(f"  Live HICON Handles: BIG=0x{hicon_big:X}, SMALL=0x{hicon_small:X}")
    if not hicon_big:
        print("  FAIL: WM_GETICON returned null HICON handle")
        subprocess.call(["taskkill", "/F", "/T", "/PID", str(proc.pid)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return False
        
    img = hicon_to_image(hicon_big)
    if not img:
        print("  FAIL: Unable to extract icon bitmap from HICON")
        subprocess.call(["taskkill", "/F", "/T", "/PID", str(proc.pid)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        return False
        
    os.makedirs(".agent/scratch", exist_ok=True)
    out_img = os.path.abspath(f".agent/scratch/verified_{profile_id}.png")
    img.save(out_img)
    
    arr = np.array(img)
    r = arr[:, :, 0].astype(int)
    g = arr[:, :, 1].astype(int)
    b = arr[:, :, 2].astype(int)
    a = arr[:, :, 3].astype(int)
    visible = a > 50
    
    passed = False
    if expected_color == "red":
        red_pixels = int(np.sum(visible & (r > 150) & (g < 100) & (b < 100)))
        blurple_pixels = int(np.sum(visible & (b > 180) & (r < 120)))
        print(f"  Pixel Analysis: Red={red_pixels}, Blurple={blurple_pixels}")
        passed = (red_pixels > 200 and blurple_pixels < 50)
    elif expected_color == "yellow":
        yellow_pixels = int(np.sum(visible & (r > 180) & (g > 180) & (b < 80)))
        blurple_pixels = int(np.sum(visible & (b > 180) & (r < 120)))
        print(f"  Pixel Analysis: Yellow={yellow_pixels}, Blurple={blurple_pixels}")
        passed = (yellow_pixels > 200 and blurple_pixels < 50)
        
    subprocess.call(["taskkill", "/F", "/T", "/PID", str(proc.pid)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    subprocess.call(["taskkill", "/F", "/IM", "DiscordPTB.exe"], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    
    status_str = "PASS" if passed else "FAIL"
    print(f"  Result: {status_str} (Saved icon proof to: {out_img})")
    return passed

if __name__ == "__main__":
    red_pass = verify_profile_window("05_red", "red")
    yellow_pass = verify_profile_window("01_yellow", "yellow")
    
    print("\n" + "=" * 50)
    print("REAL MACHINE WINDOW ICON TEST SUITE")
    print(f"  Discord Red    (05_red):    {'PASS' if red_pass else 'FAIL'}")
    print(f"  Discord Yellow (01_yellow): {'PASS' if yellow_pass else 'FAIL'}")
    print("=" * 50)
    
    if not (red_pass and yellow_pass):
        sys.exit(1)
    print("ALL REAL WINDOW ICON TESTS PASSED SUCCESSFULLY.")
