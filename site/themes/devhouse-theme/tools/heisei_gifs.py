#!/usr/bin/env python3
"""Draw the Heisei theme's heading and label images (GIF only, like it's 1999).

Uses macOS's bundled Japanese fonts (Osaka, Hiragino), so run it on a Mac:

    python3 -m venv /tmp/v && /tmp/v/bin/pip install pillow
    /tmp/v/bin/python site/themes/devhouse-theme/tools/heisei_gifs.py

Trip labels and author headings come from site/hugo.toml and the posts, so
re-run it after adding a trip or a new author.
"""
import glob
import re
import tomllib
import unicodedata
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

THEME = Path(__file__).resolve().parent.parent
SITE = THEME.parent.parent
OUT = THEME / "static" / "img" / "heisei"

KEY = (255, 0, 255)  # transparent colour; nothing is antialiased so it stays pure
NAVY, BLUE, PINK, CYAN = (0, 0, 128), (16, 132, 208), (255, 20, 147), (0, 191, 255)
HOTPINK, YELLOW, RED, GREEN = (255, 105, 180), (255, 255, 0), (255, 0, 0), (0, 128, 0)
WHITE, BLACK, ORANGE = (255, 255, 255), (0, 0, 0), (255, 69, 0)

MENU_W, MAIN_W = 164, 570  # heading widths for the two table columns (see heisei.css)


def font_path(pattern):
    for root in ("/System/Library/Fonts", "/System/Library/AssetsV2", "/Library/Fonts"):
        for p in glob.glob(f"{root}/**/*", recursive=True):
            if re.search(pattern, unicodedata.normalize("NFC", p)):
                return p
    raise SystemExit(f"font not found: {pattern}")


OSAKA = font_path(r"/Osaka\.ttf$")
MARU = font_path(r"ヒラギノ丸ゴ ProN W4\.ttc$")
GOTHIC_W6 = font_path(r"ヒラギノ角ゴシック W6\.ttc$")


def font(path, size):
    return ImageFont.truetype(path, size)


def canvas(w, h, bg=KEY):
    im = Image.new("RGB", (w, h), bg)
    d = ImageDraw.Draw(im)
    d.fontmode = "1"  # no antialiasing: crunchy bitmap text
    return im, d


def text_w(d, s, f):
    l, _, r, _ = d.textbbox((0, 0), s, font=f)
    return r - l


def text(d, xy, s, f, fill, bold=False, shadow=None):
    x, y = xy
    if shadow:
        d.text((x + 1, y + 1), s, font=f, fill=shadow)
        if bold:
            d.text((x + 2, y + 1), s, font=f, fill=shadow)
    d.text((x, y), s, font=f, fill=fill)
    if bold:
        d.text((x + 1, y), s, font=f, fill=fill)


def gradient(d, w, h, a, b, x0=0, y0=0):
    for x in range(w):
        t = x / max(w - 1, 1)
        d.line([(x0 + x, y0), (x0 + x, y0 + h - 1)], fill=tuple(round(a[i] + (b[i] - a[i]) * t) for i in range(3)))


def bevel(d, w, h, fill, x0=0, y0=0, inset=False):
    hi, lo = ((128, 128, 128), WHITE) if inset else (WHITE, (128, 128, 128))
    d.rectangle([x0, y0, x0 + w - 1, y0 + h - 1], fill=fill)
    d.line([(x0, y0), (x0 + w - 1, y0)], fill=hi)
    d.line([(x0, y0), (x0, y0 + h - 1)], fill=hi)
    d.line([(x0, y0 + h - 1), (x0 + w - 1, y0 + h - 1)], fill=lo)
    d.line([(x0 + w - 1, y0), (x0 + w - 1, y0 + h - 1)], fill=lo)
    d.line([(x0 + 1, y0 + h - 2), (x0 + w - 2, y0 + h - 2)], fill=(192, 192, 192) if not inset else fill)


SPRITES = {
    "snow": ["...#...", ".#.#.#.", "..###..", "#######", "..###..", ".#.#.#.", "...#..."],
    "flower": [".##.##.", "#######", "###o###", ".#####.", "###o###", "#######", ".##.##."],
    "leaf": ["...#...", "#..#..#", "##.#.##", ".#####.", "..###..", "...#...", "...#..."],
    "diamond": ["...#...", "..###..", ".#####.", "#######", ".#####.", "..###..", "...#..."],
}


def sprite(d, name, x, y, color, scale=1, accent=YELLOW):
    for j, row in enumerate(SPRITES[name]):
        for i, c in enumerate(row):
            if c != ".":
                d.rectangle([x + i * scale, y + j * scale, x + (i + 1) * scale - 1, y + (j + 1) * scale - 1],
                            fill=accent if c == "o" else color)


def save(im, name):
    """Save with KEY as the transparent colour (if it appears)."""
    p = im.quantize(colors=256, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE)
    pal = p.getpalette()[:768]
    idx = next((i for i in range(len(pal) // 3) if tuple(pal[i * 3:i * 3 + 3]) == KEY), None)
    kw = {"transparency": idx} if idx is not None and KEY in [c for _, c in im.getcolors(1 << 16) or []] else {}
    p.save(OUT / name, **kw)


def save_anim(frames, name, ms):
    pf = [f.quantize(colors=256, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE) for f in frames]
    pf[0].save(OUT / name, save_all=True, append_images=pf[1:], duration=ms, loop=0, disposal=1)


# ---- pieces ----------------------------------------------------------------

def heading(name, label, w):
    im, d = canvas(w, 20)
    gradient(d, w, 20, NAVY, BLUE)
    text(d, (6, 4), "■" + label, font(OSAKA, 12), WHITE, bold=True, shadow=BLACK)
    for i in range(3):
        sprite(d, "snow", w - 12 - i * 11, 7 - (i % 2) * 2, WHITE if i != 1 else (180, 220, 255))
    save(im, f"h-{name}.gif")


def logo():
    title, sub = "デブハウス☆ホームページ", "～ 開発者たちの日本旅行日記 ～ since 2024.03"
    ft, fs = font(MARU, 36), font(OSAKA, 13)
    probe = ImageDraw.Draw(Image.new("RGB", (1, 1)))
    tw, sw = text_w(probe, title, ft), text_w(probe, sub, fs)
    w = max(tw + 110, sw + 20)
    im, d = canvas(w, 66)
    x = (w - tw) // 2
    for dx, c in ((4, CYAN), (3, CYAN), (2, WHITE), (1, WHITE)):
        d.text((x + dx, 2 + dx), title, font=ft, fill=c)
    d.text((x, 2), title, font=ft, fill=PINK)
    sprite(d, "snow", x - 40, 12, PINK, scale=3)
    sprite(d, "snow", x + tw + 18, 12, PINK, scale=3)
    text(d, ((w - sw) // 2, 48), sub, fs, GREEN)
    save(im, "logo.gif")


def button(name, label, w, h=22, new=False):
    f = font(OSAKA, 12)
    for state, fill, ink in (("", (255, 228, 241), NAVY), ("-on", YELLOW, RED)):
        frames = []
        for blink in ((0, 1) if new and not state else (0,)):
            im, d = canvas(w, h)
            bevel(d, w, h, fill)
            sprite(d, "diamond", 5, (h - 7) // 2, PINK)
            text(d, (16, (h - 12) // 2 - 1), label, f, ink)
            if new and not blink:
                nx = 16 + text_w(d, label, f) + 6
                d.rectangle([nx, 5, nx + 29, h - 7], fill=RED)
                text(d, (nx + 2, 4), "NEW!", font(OSAKA, 10), WHITE)
            frames.append(im)
        if len(frames) > 1:
            save_anim(frames, f"b-{name}{state}.gif", 500)
        else:
            save(frames[0], f"b-{name}{state}.gif")


def new_badge():
    frames = []
    for bg, ink in ((RED, WHITE), (YELLOW, RED)):
        im, d = canvas(30, 13, bg)
        text(d, (3, 0), "NEW", font(OSAKA, 11), ink, bold=True)
        frames.append(im)
    save_anim(frames, "new.gif", 500)


def trip_label(slug, label, season):
    f = font(OSAKA, 13)
    probe = ImageDraw.Draw(Image.new("RGB", (1, 1)))
    w = text_w(probe, label, f) + 34
    im, d = canvas(w, 20)
    d.polygon([(0, 0), (w - 8, 0), (w - 1, 10), (w - 8, 19), (0, 19)], fill=HOTPINK)
    icon = {"冬": "snow", "春": "flower"}.get(season, "leaf")
    sprite(d, icon, 6, 6, WHITE if icon == "snow" else (255, 230, 120) if icon == "leaf" else WHITE, accent=RED)
    text(d, (18, 3), label, f, WHITE, bold=True, shadow=(170, 0, 90))
    save(im, f"trip-{slug}.gif")


def koji(label):
    w, h = MAIN_W + 6, 56  # as wide as a whole box, border included
    f = font(GOTHIC_W6, 16)
    frames = []
    for i in range(4):
        im, d = canvas(w, h, YELLOW)
        for x in range(-h - 24 + i * 6, w + h, 24):
            d.polygon([(x, h), (x + 12, h), (x + 12 + h, 0), (x + h, 0)], fill=(34, 34, 34))
        tw = text_w(d, label, f)
        bx = (w - tw - 70) // 2
        d.rectangle([bx, 10, bx + tw + 70, h - 11], fill=WHITE, outline=ORANGE)
        for sx in (bx + 8, bx + tw + 48):
            sprite(d, "snow", sx, 18, CYAN if i % 2 else BLUE, scale=2)
        text(d, (bx + 35, 18), label, f, ORANGE if i < 2 else RED)
        frames.append(im)
    save_anim(frames, "koji.gif", 160)


def badge88(name, left, right, lc, lbg, rc, rbg, split=48):
    im, d = canvas(88, 31, BLACK)
    d.rectangle([1, 1, split, 29], fill=lbg)
    d.rectangle([split + 1, 1, 86, 29], fill=rbg)
    f = font(OSAKA, 10)
    text(d, (1 + (split - text_w(d, left, f)) // 2, 8), left, f, lc, bold=True)
    text(d, (split + 1 + (86 - split - text_w(d, right, f)) // 2, 8), right, f, rc, bold=True)
    save(im, f"88-{name}.gif")


def top_button():
    im, d = canvas(110, 18)
    bevel(d, 110, 18, (220, 235, 255))
    text(d, (8, 2), "▲ページの先頭へ", font(OSAKA, 12), NAVY)
    save(im, "b-pagetop.gif")


# ---- run -------------------------------------------------------------------

def main():
    OUT.mkdir(parents=True, exist_ok=True)
    cfg = tomllib.loads((SITE / "hugo.toml").read_text())
    params = cfg.get("params", {})
    names = {k.lower(): v for k, v in params.get("names", {}).items()}

    logo()
    new_badge()
    top_button()
    for n, label in (("menu", "メニュー"), ("counter", "カウンター"), ("env", "推奨環境")):
        heading(n, label, MENU_W)
    for n, label in (("aisatsu", "ごあいさつ"), ("koshin", "更新履歴"), ("mokuji", "旅日記 もくじ"),
                     ("members", "メンバー紹介"), ("nikki", "旅日記"), ("photo", "写真館"), ("error", "エラー")):
        heading(n, label, MAIN_W)

    authors = set()
    for p in (SITE / "content").glob("*.md"):
        fm = p.read_text(encoding="utf-8").split("---")[1] if p.read_text(encoding="utf-8").startswith("---") else ""
        m = re.search(r"^authors:[ \t]*(\[[^\]]*\]|(?:\n[ \t]*- .*)+)", fm, re.M)
        if m:
            authors.update(a.strip(" -'\"\t") for a in re.split(r"[,\n]", m.group(1).strip("[]")) if a.strip(" -'\"\t"))
    for a in sorted(authors):
        heading("author-" + a.lower(), f"{names.get(a.lower(), a)}の旅日記", MAIN_W)

    for n, label, new in (("top", "トップ", False), ("nikki", "旅日記", True), ("photo", "写真館", False),
                          ("members", "メンバー紹介", False), ("game", "ゲーム", False), ("rss", "RSS", False)):
        button(n, label, MENU_W - 20, new=new)
    for n, label in (("prev", "≪前の日記"), ("mokuji", "もくじに戻る"), ("next", "次の日記≫")):
        button(n, label, 110)

    trips = params.get("trips", [])
    for t in trips:
        slug = re.sub(r"[^a-z0-9]+", "-", t["name"].lower()).strip("-")
        trip_label(slug, f'{t["name"]}（{t["season"]}） {t["note"]}', t["season"])
    if trips:
        latest = max(trips, key=lambda t: t["from"])
        koji(f'{latest["name"]}  {latest["note"]}  ただいま準備中！')

    badge88("netscape", "Netscape", "NOW!", WHITE, NAVY, BLACK, (255, 204, 0), split=56)
    badge88("yukiguni", "雪国", "RING", NAVY, (200, 230, 255), WHITE, PINK, split=40)
    badge88("ie5", "IE5.0", "推奨", WHITE, (0, 102, 204), NAVY, WHITE, split=44)
    badge88("linkfree", "リンク", "フリー", WHITE, GREEN, GREEN, (220, 255, 220), split=44)

    print(f"wrote {len(list(OUT.glob('*.gif')))} gifs to {OUT}")


if __name__ == "__main__":
    main()
