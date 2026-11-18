"""Build static/images/shinkansen.png: a full N700 train for the footer animation.

The noses are Gabrielchl's CC BY 4.0 illustration (credited in the footer):
https://commons.wikimedia.org/wiki/File:Shinkansen_head_illustration_side_view.png
Middle cars are tiled from its plain body column, windows, doors and bogies,
and two carry the devhouse logo (site/static/images/logo-dark.svg rendered to PNG).

    python build_shinkansen.py head.png logo.png ../static/images/shinkansen.png [cars=6] [height=96]
"""
import sys
from PIL import Image, ImageDraw

head_path, logo_path, out_path = sys.argv[1:4]
CARS = int(sys.argv[4]) if len(sys.argv) > 4 else 6
TARGET_H = int(sys.argv[5]) if len(sys.argv) > 5 else 96

head = Image.open(head_path).convert('RGBA')
HW, HH = head.size
CAR_W = 2200            # middle car length at source scale
PERIOD = 205            # window spacing measured from the illustration

wall = head.crop((110, 0, 111, HH))           # one clean column of body, stripes and underbody
window = head.crop((162, 140, 269, 275))      # one window (with a little body around it)
door = head.crop((452, 40, 614, 432))         # the passenger door
bogie = head.crop((780, 596, 976, HH))        # wheels and the equipment box above them
logo = Image.open(logo_path).convert('RGBA')
logo = logo.resize((round(logo.width * 160 / logo.height), 160), Image.NEAREST)


def middle_car(with_logo):
    car = Image.new('RGBA', (CAR_W, HH))
    for x in range(CAR_W):
        car.alpha_composite(wall, (x, 0))
    d = ImageDraw.Draw(car)
    # doors near both ends
    car.alpha_composite(door, (60, 40))
    car.alpha_composite(door.transpose(Image.FLIP_LEFT_RIGHT), (CAR_W - 60 - door.width, 40))
    # windows between the doors, leaving room for the logo
    x0, x1 = 60 + door.width + 70, CAR_W - 60 - door.width - 70
    lx = (CAR_W - logo.width) // 2
    x = x0
    while x + window.width <= x1:
        if not (with_logo and lx - 40 < x + window.width and x < lx + logo.width + 40):
            car.alpha_composite(window, (x, 140))
        x += PERIOD
    if with_logo:
        car.alpha_composite(logo, (lx, 98))
    for bx in (300, CAR_W - 300 - bogie.width):
        car.alpha_composite(bogie, (bx, 596))
    # gangway gap at the right end
    d.rectangle([CAR_W - 10, 4, CAR_W - 1, 632], fill=(120, 124, 132, 255))
    d.rectangle([CAR_W - 7, 4, CAR_W - 4, 632], fill=(60, 62, 68, 255))
    return car


nose = head                                 # points right: the leading end
tail = head.transpose(Image.FLIP_LEFT_RIGHT)
W = HW * 2 + CAR_W * CARS
train = Image.new('RGBA', (W, HH))
train.alpha_composite(tail, (0, 0))
for i in range(CARS):
    train.alpha_composite(middle_car(with_logo=i in (1, CARS - 2)), (HW + i * CAR_W, 0))
train.alpha_composite(nose, (HW + CARS * CAR_W, 0))

scale = TARGET_H / HH
out = train.resize((round(W * scale), TARGET_H), Image.LANCZOS)
out.save(out_path, optimize=True)
print(out.size)
