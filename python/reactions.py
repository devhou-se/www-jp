from flask import Flask, Response
from flask_cors import CORS

import requests
import os

DEFAULT_PORT = 8080

app = Flask(__name__)
CORS(app)

token = os.getenv("TOKEN")


@app.route("/")
def health_check():
    return "OK"


@app.route("/<int:post_number>/reactions")
def get_reactions(post_number: int):
    resp = Response(
        requests.get("https://api.github.com/repos/{owner}/{repo}/issues/{issue_number}/reactions".format(**{
            "owner": "devhou-se",
            "repo": "www-jp",
            "issue_number": post_number,
        }), headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {token}",
            "X-GitHub-Api-Version": "2022-11-28",
        }).content
    )
    resp.headers["Content-Type"] = "application/json"
    return resp


if __name__ == '__main__':
    print("Starting server")
    print(os.getenv("PORT"))
    port = os.getenv("PORT", DEFAULT_PORT)
    app.run(host="0.0.0.0", port=port)

