#!/usr/bin/env python3
"""
Convert Discord profile PNG icons into circular (round) icons like the official Discord app,
exporting high-resolution multi-size Windows .ico files and web PNG assets.
"""

import os
import shutil
import json
from pathlib import Path
from PIL import Image, ImageDraw

BASE_DIR = Path(__file__).resolve().parent.parent
SOURCE_DIR = BASE_DIR / "discord_icons"
ICO_OUT_DIR = BASE_DIR / "res" / "icons"
WEB_ASSETS_DIR = BASE_DIR / "cmd" / "websetup" / "web" / "assets" / "icons"
METADATA_FILE = ICO_OUT_DIR / "metadata.json"

ICO_SIZES = [(16, 16), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]

COLOR_DEFS = [
    {
        "id": 1,
        "code": "01",
        "key": "yellow",
        "name_th": "สีเหลือง",
        "name_en": "Yellow",
        "hex": "#FFFF00",
        "filename_prefix": "01_yellow_FFFF00",
        "primary_rpc": True,
    },
    {
        "id": 2,
        "code": "02",
        "key": "yellow_orange",
        "name_th": "สีส้มอมเหลือง",
        "name_en": "Yellow Orange",
        "hex": "#FFC000",
        "filename_prefix": "02_yellow_orange_FFC000",
        "primary_rpc": False,
    },
    {
        "id": 3,
        "code": "03",
        "key": "orange",
        "name_th": "สีส้ม",
        "name_en": "Orange",
        "hex": "#F58220",
        "filename_prefix": "03_orange_F58220",
        "primary_rpc": False,
    },
    {
        "id": 4,
        "code": "04",
        "key": "red_orange",
        "name_th": "สีส้มอมแดง",
        "name_en": "Red Orange",
        "hex": "#FA6304",
        "filename_prefix": "04_red_orange_FA6304",
        "primary_rpc": False,
    },
    {
        "id": 5,
        "code": "05",
        "key": "red",
        "name_th": "สีแดง",
        "name_en": "Red",
        "hex": "#ED1C24",
        "filename_prefix": "05_red_ED1C24",
        "primary_rpc": False,
    },
    {
        "id": 6,
        "code": "06",
        "key": "red_violet",
        "name_th": "สีม่วงอมแดง",
        "name_en": "Red Violet",
        "hex": "#BD0074",
        "filename_prefix": "06_red_violet_BD0074",
        "primary_rpc": False,
    },
    {
        "id": 7,
        "code": "07",
        "key": "violet",
        "name_th": "สีม่วงเข้ม",
        "name_en": "Violet",
        "hex": "#59277A",
        "filename_prefix": "07_violet_59277A",
        "primary_rpc": False,
    },
    {
        "id": 8,
        "code": "08",
        "key": "blue_violet",
        "name_th": "สีน้ำเงินอมม่วง",
        "name_en": "Blue Violet",
        "hex": "#4400A3",
        "filename_prefix": "08_blue_violet_4400A3",
        "primary_rpc": False,
    },
    {
        "id": 9,
        "code": "09",
        "key": "blue",
        "name_th": "สีน้ำเงิน",
        "name_en": "Blue",
        "hex": "#0020C2",
        "filename_prefix": "09_blue_0020C2",
        "primary_rpc": False,
    },
    {
        "id": 10,
        "code": "10",
        "key": "blue_green",
        "name_th": "สีเขียวอมฟ้า",
        "name_en": "Blue Green",
        "hex": "#008240",
        "filename_prefix": "10_blue_green_008240",
        "primary_rpc": False,
    },
    {
        "id": 11,
        "code": "11",
        "key": "green",
        "name_th": "สีเขียว",
        "name_en": "Green",
        "hex": "#00B050",
        "filename_prefix": "11_green_00B050",
        "primary_rpc": False,
    },
    {
        "id": 12,
        "code": "12",
        "key": "yellow_green",
        "name_th": "สีเขียวตองอ่อน",
        "name_en": "Yellow Green",
        "hex": "#8EF226",
        "filename_prefix": "12_yellow_green_8EF226",
        "primary_rpc": False,
    },
]

def make_circular(img: Image.Image, padding: int = 4) -> Image.Image:
    """Creates a smooth, anti-aliased circular icon with transparent background."""
    w, h = img.size
    scale = 4  # 4x supersampling for buttery smooth curved edges
    big_mask = Image.new("L", (w * scale, h * scale), 0)
    draw = ImageDraw.Draw(big_mask)
    
    pad = padding * scale
    draw.ellipse((pad, pad, w * scale - pad, h * scale - pad), fill=255)
    
    # High-quality Lanczos downsampling
    mask = big_mask.resize((w, h), Image.Resampling.LANCZOS)
    
    circular_img = Image.new("RGBA", (w, h), (0, 0, 0, 0))
    circular_img.paste(img, (0, 0), mask=mask)
    return circular_img

def main():
    print(f"Creating directories:\n  {ICO_OUT_DIR}\n  {WEB_ASSETS_DIR}")
    ICO_OUT_DIR.mkdir(parents=True, exist_ok=True)
    WEB_ASSETS_DIR.mkdir(parents=True, exist_ok=True)

    metadata_list = []

    for item in COLOR_DEFS:
        src_png_name = f"{item['filename_prefix']}.png"
        src_png_path = SOURCE_DIR / src_png_name
        
        if not src_png_path.exists():
            raise FileNotFoundError(f"Source icon not found: {src_png_path}")

        print(f"Processing Circular Icon {item['id']:02d}: {item['name_en']} ({item['hex']})...")

        with Image.open(src_png_path) as img:
            rgba_img = img.convert("RGBA")
            
            # Apply circular mask (perfect circle like Discord official app)
            circular_img = make_circular(rgba_img, padding=4)

            # Overwrite source PNG with clean circular PNG
            circular_img.save(src_png_path, format="PNG")

            # Save multi-size .ico in res/icons/
            primary_ico_path = ICO_OUT_DIR / f"{item['filename_prefix']}.ico"
            circular_img.save(primary_ico_path, format="ICO", sizes=ICO_SIZES)

            # Also create alias profile_X.ico and profile_XX.ico for easy lookup
            alias_ico_num = ICO_OUT_DIR / f"profile_{item['id']}.ico"
            alias_ico_code = ICO_OUT_DIR / f"profile_{item['code']}.ico"
            shutil.copy2(primary_ico_path, alias_ico_num)
            shutil.copy2(primary_ico_path, alias_ico_code)

            # Save circular PNG to web assets
            dest_web_png = WEB_ASSETS_DIR / src_png_name
            circular_img.save(dest_web_png, format="PNG")

            # Also copy alias profile_XX.png in web assets
            alias_web_png = WEB_ASSETS_DIR / f"profile_{item['code']}.png"
            shutil.copy2(dest_web_png, alias_web_png)

        entry = {
            "id": item["id"],
            "code": item["code"],
            "key": item["key"],
            "name_th": item["name_th"],
            "name_en": item["name_en"],
            "hex": item["hex"],
            "primary_rpc": item["primary_rpc"],
            "filename_base": item["filename_prefix"],
            "ico_file": f"{item['filename_prefix']}.ico",
            "png_file": src_png_name,
            "alias_ico": f"profile_{item['id']}.ico",
            "alias_png": f"profile_{item['code']}.png",
            "shape": "circle",
        }
        metadata_list.append(entry)

    # Update app default icons (Profile 1 Yellow as circular .ico)
    primary_yellow_ico = ICO_OUT_DIR / "01_yellow_FFFF00.ico"
    if primary_yellow_ico.exists():
        shutil.copy2(primary_yellow_ico, BASE_DIR / "res" / "papp.ico")
        shutil.copy2(primary_yellow_ico, BASE_DIR / "res" / "app.ico")

    # Save metadata.json
    output_metadata = {
        "version": "1.1.0",
        "shape": "circular",
        "count": len(metadata_list),
        "profiles": metadata_list,
    }

    with open(METADATA_FILE, "w", encoding="utf-8") as f:
        json.dump(output_metadata, f, indent=2, ensure_ascii=False)
    print(f"Metadata saved to: {METADATA_FILE}")

    web_metadata_file = WEB_ASSETS_DIR / "metadata.json"
    with open(web_metadata_file, "w", encoding="utf-8") as f:
        json.dump(output_metadata, f, indent=2, ensure_ascii=False)
    print(f"Web metadata saved to: {web_metadata_file}")

    print("\nAll 12 Discord icons have been converted into circular icons successfully!")

if __name__ == "__main__":
    main()
