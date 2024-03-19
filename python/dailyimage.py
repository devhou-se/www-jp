import os
import random
import re

GALLERY_MD = os.path.join(os.getcwd(), "site", "content", "gallery.md")
GALLERY_HTML = os.path.join(os.getcwd(), "site", "themes", "devhouse-theme", "layouts", "partials", "gallery.html")
IMG_PATTERN = r"(\!\[(.*)\]\((.*)\))"


def main():
    with open(GALLERY_MD, "r") as f:
        lines = f.read().split("\n")

    rand_line = random.choice(lines[4:])
    img = re.search(IMG_PATTERN, rand_line)
    alt, uri = img.group(2), img.group(3)

    dom = f"<img src=\"{uri}\" alt=\"{alt}\">"

    with open(GALLERY_HTML, "w") as f:
        f.write(dom)


if __name__ == '__main__':
    main()
