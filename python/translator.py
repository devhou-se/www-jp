import json

from openai import OpenAI

from enum import Enum
import os
import sys
import time

OPENAI_KEY = os.getenv("OPENAI_API_KEY")
ASSISTANT_ID = "asst_IK9udKO5TJSVcCyBoL56TdiE"


def main():
    filename = f"{sys.argv[1]}.md"
    full_filename = os.path.join(os.getcwd(), "site", "content", filename)

    with open(full_filename, "r") as f:
        content = f.read()

    jp = analyze_text(content)
    if jp > 0.5:
        return

    translated_text = translate_md(content)

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


def translate_md(inp: str) -> str:
    client = OpenAI(api_key=OPENAI_KEY)
    thread = client.beta.threads.create()
    _ = client.beta.threads.messages.create(thread.id, role="user", content=inp)
    run = client.beta.threads.runs.create(thread.id, assistant_id=ASSISTANT_ID, model="gpt-4")

    while run.status in ["queued", "in_progress", "cancelling"]:
        print(f"waiting for run {run.id} to complete. status: {run.status}")
        time.sleep(1)
        run = client.beta.threads.runs.retrieve(
            thread_id=thread.id,
            run_id=run.id
        )

    if run.status == "completed":
        messages = client.beta.threads.messages.list(
            thread_id=thread.id
        )

        print(messages.data[0].content[0].text.value)

        return messages.data[0].content[0].text.value

    else:
        print(f"run {run.id} failed with status: {run.status}")
        sys.exit(0)


if __name__ == '__main__':
    main()
