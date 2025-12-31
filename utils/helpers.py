import os
import json

ALLOWED_EXTENSIONS = {"zip", "pdf", "md"}
HOMEWORKS = []

with open(os.path.join(os.path.dirname(__file__), "..", "homeworks.json"), "r", encoding="utf-8") as f:
    HOMEWORKS = json.load(f)


def allowed_file(filename):
    return (
        "." in filename and filename.rsplit(
            ".", 1)[1].lower() in ALLOWED_EXTENSIONS
    )


def secure_filename(filename):
    return filename.replace("/", "").replace("\\", "").replace("..", "")


def get_homework_name(num: int):
    global HOMEWORKS
    names = {i["id"]: i["name"] for i in HOMEWORKS}
    return names.get(num, f"第{num}次作业")
