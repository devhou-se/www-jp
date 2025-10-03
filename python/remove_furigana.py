import os
import re

def remove_furigana(text: str) -> str:
    """Remove furigana annotations from text, keeping only the kanji."""
    return re.sub(r'\[([^\]]*)\]\{[^\}]*\}', r'\1', text)

def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")

    for filename in os.listdir(content_dir):
        if not filename.endswith('.md'):
            continue

        filepath = os.path.join(content_dir, filename)

        with open(filepath, "r") as f:
            content = f.read()

        # Check if file has furigana annotations
        if re.search(r'\[([^\]]*)\]\{[^\}]*\}', content):
            cleaned_content = remove_furigana(content)

            with open(filepath, "w") as f:
                f.write(cleaned_content)

            print(f"Cleaned {filename}")

if __name__ == '__main__':
    main()
