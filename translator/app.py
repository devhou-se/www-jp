import openai
import os

print(os.getcwd())

# iter dir
for root, dirs, files in os.walk("."):
    for file in files:
        print(file)
