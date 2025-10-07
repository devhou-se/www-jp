from google.cloud import translate_v3
import yaml

from enum import Enum
import argparse
import os
import sys

PROJECT_ID = os.getenv("GOOGLE_CLOUD_PROJECT")
LOCATION = "global"


def separate_frontmatter(content: str) -> tuple[dict, str]:
    """Separate YAML frontmatter from markdown body."""
    if not content.startswith("---"):
        return {}, content

    # Find the end of frontmatter
    lines = content.split("\n")
    end_index = -1
    for i in range(1, len(lines)):
        if lines[i].strip() == "---":
            end_index = i
            break

    if end_index == -1:
        return {}, content

    frontmatter_text = "\n".join(lines[1:end_index])
    body = "\n".join(lines[end_index + 1:]).strip()

    # Parse YAML frontmatter
    try:
        frontmatter = yaml.safe_load(frontmatter_text)
        if frontmatter is None:
            frontmatter = {}
    except yaml.YAMLError:
        frontmatter = {}

    return frontmatter, body


def rebuild_frontmatter(frontmatter: dict) -> str:
    """Rebuild YAML frontmatter with proper formatting."""
    if not frontmatter:
        return ""

    yaml_content = yaml.dump(frontmatter, allow_unicode=True, default_flow_style=False, sort_keys=False)
    return f"---\n{yaml_content}---"


def main():
    parser = argparse.ArgumentParser(description="Translate markdown files")
    parser.add_argument("issue_number", help="Issue number for the markdown file")
    parser.add_argument("--to-lang", default="ja", help="Target language code (default: ja)")
    args = parser.parse_args()

    target_lang = args.to_lang

    filename = f"{args.issue_number}.md"
    full_filename = os.path.join(os.getcwd(), "site", "content", filename)

    with open(full_filename, "r") as f:
        content = f.read()

    jp = analyze_text(content)
    print(f"Japanese text score: {(jp*100):.2f}%")
    if jp > 0.5:
        return

    # Separate frontmatter from content
    frontmatter, body = separate_frontmatter(content)

    # Translate the title in frontmatter if it exists
    if frontmatter and "title" in frontmatter:
        frontmatter["title"] = translate_md(frontmatter["title"], target_lang)

    # Translate the body content
    translated_body = translate_md(body, target_lang)

    # Recombine with translated frontmatter
    frontmatter_text = rebuild_frontmatter(frontmatter)
    translated_text = f"{frontmatter_text}\n{translated_body}" if frontmatter_text else translated_body

    with open(full_filename, "w") as f:
        f.write(translated_text)


class Script(Enum):
    LATIN = "latin"
    JAPANESE = "japanese"
    OTHER = "other"


def categorize_character(char):
    code_point = ord(char)

    # Latin script ranges (including extended)
    if (0x0041 <= code_point <= 0x005A) or (0x0061 <= code_point <= 0x007A) or \
            (0x00C0 <= code_point <= 0x00FF) or (0x0100 <= code_point <= 0x024F):
        return Script.LATIN

    # Japanese script ranges (simplified, includes common Hiragana, Katakana, and Kanji)
    if (0x3040 <= code_point <= 0x309F) or \
            (0x30A0 <= code_point <= 0x30FF) or \
            (0x4E00 <= code_point <= 0x9FAF) or \
            (0x3400 <= code_point <= 0x4DBF):  # Additional range for CJK Unified Ideographs Extension
        return Script.JAPANESE

    return Script.OTHER


def analyze_text(t: str) -> float:
    script_counts = {k: 0 for k in Script}
    for char in t:
        category = categorize_character(char)
        script_counts[category] += 1

    return script_counts[Script.JAPANESE] / (script_counts[Script.JAPANESE] + script_counts[Script.LATIN])


def translate_md(inp: str, target_lang: str = "ja") -> str:
    """Translate markdown text from English to target language using Google Cloud Translation API."""
    client = translate_v3.TranslationServiceClient()
    parent = f"projects/{PROJECT_ID}/locations/{LOCATION}"

    print(f"Translating text using project: {PROJECT_ID} to language: {target_lang}")

    response = client.translate_text(
        contents=[inp],
        parent=parent,
        mime_type="text/plain",
        source_language_code="en",
        target_language_code=target_lang,
    )

    if response.translations:
        translated_text = response.translations[0].translated_text
        print(f"Translation completed successfully")
        return translated_text
    else:
        print(f"Translation failed: no translations returned")
        sys.exit(1)


if __name__ == '__main__':
    main()
