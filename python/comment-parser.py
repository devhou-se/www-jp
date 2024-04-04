import sys

data = sys.stdin.read()
data = data.split('\n')[0]
data = data.split(" ")

if data[0] == "/post":
    print("::set-output name=valid::true")
else:
    print("::set-output name=valid::false")

if len(data) > 1 and data[1] == "--no-translate":
    print("::set-output name=translate::false")
else:
    print("::set-output name=translate::true")
