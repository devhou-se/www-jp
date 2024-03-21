import datetime
import os
import re

MAX_AGE = 2  # days

GALLERY_MD = os.path.join(os.getcwd(), "site", "content", "gallery.md")
GALLERY_HTML = os.path.join(os.getcwd(), "site", "themes", "devhouse-theme", "layouts", "partials", "gallery.html")
IMG_PATTERN = r"(\!\[(.*)\]\((.*)\))"

IMG_TEMPLATE = "<img src=\"{}\" alt=\"{}\" />"
IMG_MD_TEMPLATE = "![{}]({})"
HTML_TEMPLATE = """<span id="daily-image"></span>
<script>
const choices = {choices};
const choice = choices[Math.floor(Math.random() * choices.length)];
const imgSpan = document.getElementById("daily-image");
imgSpan.innerHTML = choice;
</script>
"""


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    all_images = []
    valid_images = []

    for filename in os.listdir(content_dir):
        # ignore gallery.md
        if filename == "gallery.md":
            continue

        if filename.endswith(".md"):
            full_filename = os.path.join(content_dir, filename)
            with open(full_filename, "r") as f:
                content = f.read()
            images, valid = collect_images(content)
            all_images += images
            if valid:
                valid_images += images

    if len(valid_images) == 0:
        print("No images found")
        with open(GALLERY_HTML, "w") as f:
            f.write("最近の写真はありません")
    else:
        images = [IMG_TEMPLATE.format(img[2], img[1]) for img in valid_images]
        images = "['" + "', '".join(images) + "']"
        js = HTML_TEMPLATE.format(choices=images)

        with open(GALLERY_HTML, "w") as f:
            f.write(js)

    with open(GALLERY_MD, "w") as f:
        f.write("""---
title: 写真ギャラリー
---
""" + "\n".join([IMG_MD_TEMPLATE.format(img[1], img[2]) for img in valid_images]))


def collect_images(data: str) -> tuple[list[str], bool]:
    if not data.startswith("---"):
        return [], False

    end = data[3:].index("---")
    meta = data[3:end].split("\n")
    kv = {}
    for line in meta:
        if not line:
            continue
        key, value = line.split(": ")
        kv[key] = value

    imgs = re.findall(IMG_PATTERN, data)

    if not post_validator(kv):
        return imgs, False

    return imgs, True


def post_validator(meta: dict[str, str]) -> bool:
    if meta.get("draft") == "true":
        return False

    if meta.get("date"):
        date = datetime.datetime.strptime(meta.get("date"), "%Y-%m-%dT%H:%M:%SZ")

        if (datetime.datetime.now() - date).days >= MAX_AGE:
            return False

    return True


if __name__ == '__main__':
    main()
