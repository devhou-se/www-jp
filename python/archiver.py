import os

POST_NUMBER = os.getenv("POST_NUMBER")

DRAFT_TRUE = "draft: true"
DRAFT_FALSE = "draft: false"


def main():
    content_dir = os.path.join(os.getcwd(), "site", "content")
    md_filename = f"{POST_NUMBER}.md"

    with open(os.path.join(content_dir, md_filename), "r") as f:
        content = f.read()

    lines = content.split("\n")
    frontmatter = []

    i = 0
    if lines[0] == "---":
        for i, line in enumerate(lines[1:]):
            if line == "---":
                break
            frontmatter.append(line)

    for line in frontmatter:
        if line.startswith("draft:"):
            frontmatter.remove(line)

    frontmatter.append(DRAFT_TRUE)
    frontmatter = ["---"] + frontmatter + ["---"]
    content = "\n".join(frontmatter) + "\n" + "\n".join(lines[i+2:])

    with open(os.path.join(content_dir, md_filename), "w") as f:
        f.write(content)


if __name__ == '__main__':
    main()
