import datetime
import os
import re
import yaml
import urllib.request

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

def download_and_rehost_image(url: str, post_number: str) -> str:
    """Download image from user-attachments and save to static folder."""
    if 'user-attachments/assets/' not in url:
        return url

    # Extract image ID from URL
    import hashlib
    img_hash = hashlib.md5(url.encode()).hexdigest()[:8]

    # Create posts images directory
    posts_dir = os.path.join(os.getcwd(), "site", "static", "images", "posts")
    os.makedirs(posts_dir, exist_ok=True)

    # Determine file extension (default to png)
    ext = 'png'

    # Download the image
    img_path = os.path.join(posts_dir, f"{post_number}-{img_hash}.{ext}")

    try:
        urllib.request.urlretrieve(url, img_path)
        print(f"Downloaded image from {url} to {img_path}")
        return f"/images/posts/{post_number}-{img_hash}.{ext}"
    except Exception as e:
        print(f"Failed to download image from {url}: {e}")
        return url

def convert_html_images_to_markdown(content: str, post_number: str) -> str:
    """Convert HTML img tags to markdown image syntax and rehost user-attachments images."""
    # Pattern to match HTML img tags with various attributes
    # Matches: <img width="..." height="..." alt="..." src="..." />
    img_pattern = r'<img\s+[^>]*?alt="([^"]*)"[^>]*?src="([^"]*)"[^>]*?/?>'

    def replace_img(match):
        alt_text = match.group(1)
        src_url = match.group(2)
        # Download and rehost if it's a user-attachments URL
        src_url = download_and_rehost_image(src_url, post_number)
        return f"![{alt_text}]({src_url})"

    # Replace all HTML img tags with markdown format
    content = re.sub(img_pattern, replace_img, content)

    # Also handle cases where alt comes after src
    img_pattern_alt_after = r'<img\s+[^>]*?src="([^"]*)"[^>]*?alt="([^"]*)"[^>]*?/?>'

    def replace_img_alt_after(match):
        src_url = match.group(1)
        alt_text = match.group(2)
        # Download and rehost if it's a user-attachments URL
        src_url = download_and_rehost_image(src_url, post_number)
        return f"![{alt_text}]({src_url})"

    content = re.sub(img_pattern_alt_after, replace_img_alt_after, content)

    # Also handle markdown images with user-attachments URLs
    md_img_pattern = r'!\[([^\]]*)\]\((https://github\.com/user-attachments/[^\)]+)\)'

    def replace_md_img(match):
        alt_text = match.group(1)
        src_url = match.group(2)
        src_url = download_and_rehost_image(src_url, post_number)
        return f"![{alt_text}]({src_url})"

    content = re.sub(md_img_pattern, replace_md_img, content)

    return content

def download_github_avatar(username: str, avatars_dir: str):
    """Download GitHub avatar for the specified username."""
    avatar_url = f"https://github.com/{username}.png?size=128"
    avatar_path = os.path.join(avatars_dir, f"{username}.png")

    # Skip if avatar already exists
    if os.path.exists(avatar_path):
        return

    try:
        urllib.request.urlretrieve(avatar_url, avatar_path)
        print(f"Downloaded avatar for {username}")
    except Exception as e:
        print(f"Failed to download avatar for {username}: {e}")

def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")
    avatars_dir = os.path.join(os.getcwd(), "site", "static", "images", "avatars")
    md_filename = f"{POST_NUMBER}.md"

    # Create avatars directory if it doesn't exist
    os.makedirs(avatars_dir, exist_ok=True)

    header, content_body = extract_header(POST_BODY)

    # Convert HTML img tags to markdown format and rehost user-attachments images
    content_body = convert_html_images_to_markdown(content_body, POST_NUMBER)

    if not header:
        frontmatter_data = {
            'title': POST_TITLE,
            'date': convert_date(POST_DATE),
            'authors': [POST_AUTHOR]
        }
        yaml_content = yaml.dump(frontmatter_data, allow_unicode=True, default_flow_style=False, sort_keys=False)
        header = f"---\n{yaml_content}---\n"

    content = f"{header}{content_body}"

    with open(os.path.join(content_dir, md_filename), "w") as f:
        f.write(content)

    # Download GitHub avatar for the author
    download_github_avatar(POST_AUTHOR, avatars_dir)

if __name__ == '__main__':
    main()
