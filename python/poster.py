import datetime
import os

POST_TITLE = os.getenv("POST_TITLE")
POST_BODY = os.getenv("POST_BODY")
POST_NUMBER = os.getenv("POST_NUMBER")
POST_AUTHOR = os.getenv("POST_AUTHOR")
POST_DATE = os.getenv("POST_DATE")

HEADER_DELIMITER = "---"

def convert_date(date: str) -> str:
    if not date[-1] == "Z":
        return date

    timestamp = datetime.datetime.fromisoformat(date[:-1])
    timestamp = timestamp.astimezone(datetime.timezone(datetime.timedelta(hours=9)))
    return timestamp.isoformat()

def extract_header(body: str):
    lines = body.split('\n')
    if lines[0].strip() == HEADER_DELIMITER and HEADER_DELIMITER in lines[1:]:
        end_index = lines[1:].index(HEADER_DELIMITER) + 1
        return '\n'.join(lines[:end_index + 1]), '\n'.join(lines[end_index + 1:])
    return None, body

def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")
    md_filename = f"{POST_NUMBER}.md"

    header, content_body = extract_header(POST_BODY)
    
    if not header:
        header = f"""---
title: {POST_TITLE}
date: {convert_date(POST_DATE)}
authors: [{POST_AUTHOR}]
---"""

    content = f"{header}\n{content_body}"

    with open(os.path.join(content_dir, md_filename), "w") as f:
        f.write(content)

if __name__ == '__main__':
    main()
