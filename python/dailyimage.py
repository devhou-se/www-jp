from enum import Enum

import datetime
import os
import re

MAX_AGE = 2  # days

GALLERY_HTML = os.path.join(os.getcwd(), "site", "themes", "devhouse-theme", "layouts", "partials", "gallery.html")
IMAGES_HTML = os.path.join(os.getcwd(), "site", "themes", "devhouse-theme", "layouts", "partials", "images.html")
IMG_PATTERN = r"(\!\[(.*)\]\((.*)\))"

IMG_TEMPLATE = "<a href=\"{}\"><img src=\"{}\" alt=\"{}\" /></a>"
IMG_MD_TEMPLATE = "![{}]({})"
HTML_TEMPLATE = """<span id="daily-image"></span>
<script>
const choices = {choices};
const choice = choices[Math.floor(Math.random() * choices.length)];
const imgSpan = document.getElementById("daily-image");
imgSpan.innerHTML = choice;
</script>
"""


class Validity(Enum):
    VALID = "Valid"
    DRAFT = "Draft"
    OLD = "Old"
    MISFORMATTED = "Misformatted"


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    images_validity = {v: [] for v in Validity}

    for filename in os.listdir(content_dir):
        if filename == "gallery.md":
            continue

        if filename.endswith(".md"):
            full_filename = os.path.join(content_dir, filename)
            with open(full_filename, "r") as f:
                content = f.read()
            images, valid = collect_images(content)
            images = [(img, filename.removesuffix(".md")) for img in images]
            images_validity[valid].extend(images)

    sidebar_images = images_validity[Validity.VALID]
    gallery_images = images_validity[Validity.VALID] + images_validity[Validity.OLD]

    if len(sidebar_images) == 0:
        print("No images found")
        with open(GALLERY_HTML, "w") as f:
            f.write("最近の写真はありません")
    else:
        images = [IMG_TEMPLATE.format(img[1], img[0][2], img[0][1]) for img in sidebar_images]
        images = "['" + "', '".join(images) + "']"
        js = HTML_TEMPLATE.format(choices=images)

        with open(GALLERY_HTML, "w") as f:
            f.write(js)

    with open(IMAGES_HTML, "w") as f:
        f.write("\n".join([IMG_TEMPLATE.format(img[1], img[0][2], img[0][1]) for img in gallery_images]))


def collect_images(data: str) -> tuple[list[str], Validity]:
    if not data.startswith("---"):
        return [], Validity.MISFORMATTED

    end = data[3:].index("---")
    meta = data[3:end+3].split("\n")
    kv = {}
    for line in meta:
        if not line:
            continue
        key, value = line.split(": ")
        kv[key] = value

    imgs = re.findall(IMG_PATTERN, data)

    return imgs, post_validator(kv)


def post_validator(meta: dict[str, str]) -> Validity:
    if meta.get("draft") == "true":
        return Validity.DRAFT

    if meta.get("date"):
        date = datetime.datetime.fromisoformat(meta.get("date"))
        tzinfo = datetime.timezone(datetime.timedelta(hours=9))
        if (datetime.datetime.now().astimezone(tzinfo) - date).days >= MAX_AGE:
            return Validity.OLD

    return Validity.VALID


if __name__ == '__main__':
    main()
