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
    You are a helpful assistant that adds furigana to Japanese text. 
    Add furigana to ALL kanji characters in the format [kanji]{furigana}.
    For example: [漢字]{かんじ} 
    Do not modify any markdown formatting, only add furigana annotations.
    Return the ENTIRE document with furigana added.
    Modify ONLY the content of the document, not the metadata or title.
    DO NOT modify the title or author.
    
    Another example:
    ```md
    ---
    title: 個人記録
    date: 2024-03-27T07:39:15+09:00
    author: baely
    ---
    [漢字]{かんじ} 
    ```
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