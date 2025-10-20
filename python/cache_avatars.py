#!/usr/bin/env python3
"""Cache GitHub avatars for all existing blog post authors."""

import os
import yaml
import urllib.request
from pathlib import Path

def extract_authors_from_posts():
    """Extract all unique authors from blog posts."""
    authors = set()
    content_dir = Path("site/content")

    for md_file in content_dir.glob("*.md"):
        if md_file.name == "gallery.md":
            continue

        try:
            with open(md_file, 'r', encoding='utf-8') as f:
                content = f.read()

            # Extract frontmatter
            if content.startswith('---'):
                parts = content.split('---', 2)
                if len(parts) >= 3:
                    frontmatter = yaml.safe_load(parts[1])
                    if frontmatter and 'authors' in frontmatter:
                        author_list = frontmatter['authors']
                        if isinstance(author_list, list):
                            authors.update(author_list)
        except Exception as e:
            print(f"Error processing {md_file}: {e}")

    return authors

def download_avatar(username, avatars_dir, force=False):
    """Download GitHub avatar for a username."""
    avatar_url = f"https://github.com/{username}.png?size=128"
    avatar_path = os.path.join(avatars_dir, f"{username}.png")

    if os.path.exists(avatar_path) and not force:
        print(f"Avatar for {username} already exists, skipping")
        return

    try:
        print(f"Downloading avatar for {username}...")
        urllib.request.urlretrieve(avatar_url, avatar_path)
        print(f"✓ Downloaded avatar for {username}")
    except Exception as e:
        print(f"✗ Failed to download avatar for {username}: {e}")

def main():
    # Create avatars directory
    avatars_dir = "site/static/images/avatars"
    os.makedirs(avatars_dir, exist_ok=True)

    # Get all unique authors
    authors = extract_authors_from_posts()
    print(f"Found {len(authors)} unique authors: {sorted(authors)}\n")

    # Download avatars (force=True to re-download at new size)
    for author in sorted(authors):
        download_avatar(author, avatars_dir, force=True)

    print(f"\nDone! Cached {len(authors)} avatars in {avatars_dir}")

if __name__ == "__main__":
    main()
