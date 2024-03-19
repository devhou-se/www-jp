import os

POST_TITLE = os.getenv("POST_TITLE")
POST_BODY = os.getenv("POST_BODY")
POST_NUMBER = os.getenv("POST_NUMBER")
POST_AUTHOR = os.getenv("POST_AUTHOR")
POST_DATE = os.getenv("POST_DATE")

BODY_FORMAT = """---
title: {title}
date: {date}
author: {author}
---
{body}
"""


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    md_filename = f"{POST_NUMBER}.md"

    content = BODY_FORMAT.format(
        title=POST_TITLE,
        date=POST_DATE,
        author=POST_AUTHOR,
        body=POST_BODY,
    )

    with open(os.path.join(content_dir, md_filename), "w") as f:
        f.write(content)


if __name__ == '__main__':
    main()
