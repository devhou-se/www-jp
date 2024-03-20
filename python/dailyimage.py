import datetime
import os
import random
import re

MAX_AGE = 2 # days

GALLERY_MD = os.path.join(os.getcwd(), "site", "content", "gallery.md")
GALLERY_HTML = os.path.join(os.getcwd(), "site", "themes", "devhouse-theme", "layouts", "partials", "gallery.html")
IMG_PATTERN = r"(\!\[(.*)\]\((.*)\))"

IMG_TEMPLATE = "<img src=\"{}\" alt=\"{}\" />"
HTML_TEMPLATE = """
<span id="daily-image"></span>
<script>
const choices = {choices};
const choice = choices[Math.floor(Math.random() * choices.length)];
const imgSpan = document.getElementById("daily-image");
imgSpan.innerHTML = choice;
</script>
"""


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    images = []

    for filename in os.listdir(content_dir):
        if filename.endswith(".md"):
            full_filename = os.path.join(content_dir, filename)
            with open(full_filename, "r") as f:
                content = f.read()
            images += collect_images(content)

    if len(images) == 0:
        print("No images found")
        with open(GALLERY_HTML, "w") as f:
            f.write("最近の画像はありません")
        return

    images = [IMG_TEMPLATE.format(img[2], img[1]) for img in images]
    images = "[\"" + "\", \"".join(images) + "\"]"
    js = HTML_TEMPLATE.format(choices=images)

    with open(GALLERY_HTML, "w") as f:
        f.write(js)


def collect_images(data: str) -> list[str]:
    if not data.startswith("---"):
        return []

    end = data[3:].index("---")
    meta = data[3:end].split("\n")
    kv = {}
    for line in meta:
        if not line:
            continue
        key, value = line.split(": ")
        kv[key] = value

    if not post_validator(kv):
        return []

    imgs = re.findall(IMG_PATTERN, data)
    return imgs


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
