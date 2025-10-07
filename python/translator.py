from google.cloud import translate_v3

from enum import Enum
import argparse
import os
import sys

PROJECT_ID = os.getenv("GOOGLE_CLOUD_PROJECT")
LOCATION = "global"


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

    translated_text = translate_md(content, target_lang)

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
