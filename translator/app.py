import json

from openai import OpenAI

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

    translated_text = translate_md(content)

    with open(full_filename, "w") as f:
        f.write(translated_text)


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

        print(messages.data)

        return messages[0].content

    else:
        print(run.status)
        sys.exit(1)


if __name__ == '__main__':
    main()
