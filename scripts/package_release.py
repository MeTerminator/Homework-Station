"""Package the native backend and built frontend without local configuration/data."""

import os
import shutil
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def main() -> None:
    tag = os.environ["RELEASE_TAG"]
    target = os.environ["RELEASE_TARGET"]
    extension = ".exe" if target == "windows-amd64" else ""
    name = f"homework-station-{tag}-{target}"
    output = ROOT / "release"
    output.mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory() as directory:
        staging = Path(directory) / name
        binary_dir = staging / "backend/bin"
        binary_dir.mkdir(parents=True)
        binary = f"homework-station{extension}"
        shutil.copy2(ROOT / "backend/bin" / binary, binary_dir / binary)
        shutil.copytree(ROOT / "web/dist", staging / "web/dist")
        for filename in ("README.md", ".env.example"):
            shutil.copy2(ROOT / filename, staging / filename)
        (staging / "VERSION").write_text(tag + "\n", encoding="utf-8")
        archive = shutil.make_archive(
            str(output / name),
            "zip" if extension else "gztar",
            root_dir=directory,
            base_dir=name,
        )
    print(archive)


if __name__ == "__main__":
    main()
