"""Copy the self-contained executable to its versioned Release asset name."""

import os
import shutil
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def main() -> None:
    tag = os.environ["RELEASE_TAG"]
    target = os.environ["RELEASE_TARGET"]
    extension = ".exe" if target == "windows-amd64" else ""
    name = f"homework-station-{tag}-{target}{extension}"
    output = ROOT / "release"
    output.mkdir(exist_ok=True)
    binary = f"homework-station{extension}"
    destination = output / name
    shutil.copy2(ROOT / "backend/bin" / binary, destination)
    print(destination)


if __name__ == "__main__":
    main()
