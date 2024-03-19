import os
import re

IMG_PATTERN = r"(\!\[.*\]\(.*\))"
MD_BASE = """---
title: All images
draft: true
---
"""


def collect_images(data: str) -> list[str]:
    if data.startswith("---"):
        _, meta, data = data.split("---", 2)
        if "draft: true" in meta:
            return []

    imgs = re.findall(IMG_PATTERN, data)
    return imgs


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    images = []

    for filename in os.listdir(content_dir):
        if filename.endswith(".md"):
            full_filename = os.path.join(content_dir, filename)
            with open(full_filename, "r") as f:
                content = f.read()
            images += collect_images(content)

    image_data = "\n".join(images)

    with open(os.path.join(content_dir, "gallery.md"), "w") as f:
        f.write(MD_BASE + image_data)


if __name__ == '__main__':
    main()
