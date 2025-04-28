import json
import os
import sys
import time

from openai import OpenAI

OPENAI_KEY = os.getenv("OPENAI_API_KEY")
MODEL = "gpt-4.1"

def main():
    filename = f"{sys.argv[1]}.md"
    full_filename = os.path.join(os.getcwd(), "site", "content", filename)

    with open(full_filename, "r") as f:
        content = f.read()

    annotated_text = add_furigana(content)

    with open(full_filename, "w") as f:
        f.write(annotated_text)


def add_furigana(text: str) -> str:
    client = OpenAI(api_key=OPENAI_KEY)
    
    system_prompt = """
# Furigana Annotation Assistant

You are a specialized assistant that adds furigana readings to Japanese text. Your sole purpose is to annotate kanji characters with their readings while preserving the original document structure.

## Annotation Format
- Add furigana to ALL kanji characters using the format: [kanji]{furigana}
- Example: [漢字]{かんじ}, [日本語]{にほんご}, [勉強]{べんきょう}
- Do NOT add furigana to hiragana or katakana characters
- For compound kanji words, provide the reading for the entire word: [日本]{にほん} not [日]{に}[本]{ほん}

## Document Handling Rules
- Preserve ALL original formatting including markdown, HTML, code blocks, and line breaks
- Return the ENTIRE document with furigana added
- Do NOT modify document metadata, titles, authors, dates, or any non-Japanese text
- If uncertain about a reading, use the most common reading for the context

## Examples:

Input:
```md
---
title: 個人記録
date: 2024-03-27T07:39:15+09:00
author: baely
---
今日は東京で友達と会いました。とても楽しかったです。
Output:
md---
title: 個人記録
date: 2024-03-27T07:39:15+09:00
author: baely
---
[今日]{きょう}は[東京]{とうきょう}で[友達]{ともだち}と[会]{あ}いました。とても[楽]{たの}しかったです。
For mixed text, only annotate the Japanese kanji:
"The Japanese word 漢字 means Chinese characters" → "The Japanese word [漢字]{かんじ} means Chinese characters"
"""
    
    messages = [
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": text}
    ]
    
    response = client.chat.completions.create(
        model=MODEL,
        messages=messages,
        temperature=0.2
    )

    print(response.usage)
    
    return response.choices[0].message.content


if __name__ == '__main__':
    main()