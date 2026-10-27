"""Format every supported source/configuration file, or check without changing files."""

import argparse
import os
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def run(command: list[str], cwd: Path = ROOT, listed_files_fail: bool = False) -> bool:
    print("+ " + " ".join(command), flush=True)
    result = subprocess.run(command, cwd=cwd, text=True, capture_output=listed_files_fail)
    if listed_files_fail and result.stdout:
        print(result.stdout, end="")
    if listed_files_fail and result.stderr:
        print(result.stderr, end="", file=sys.stderr)
    return result.returncode == 0 and not (listed_files_fail and result.stdout.strip())


def go_module(check: bool) -> bool:
    result = subprocess.run(
        ["go", "mod", "edit", "-fmt", "-print"],
        cwd=ROOT / "backend",
        text=True,
        capture_output=True,
    )
    if result.returncode != 0:
        print(result.stderr, file=sys.stderr)
        return False
    module = ROOT / "backend/go.mod"
    if check:
        if module.read_text() != result.stdout:
            print("Go module formatting differs: backend/go.mod")
            return False
    else:
        module.write_text(result.stdout)
    return True


def whitespace(check: bool) -> bool:
    """Keep Makefile recipes intact while normalizing line endings and trailing spaces."""
    files = [
        "Makefile",
        "Dockerfile",
        ".editorconfig",
        ".gitignore",
        ".dockerignore",
        ".prettierignore",
        ".env.example",
        "requirements-dev.txt",
        "nonebot_plugin_homework/requirements.txt",
        "nonebot_plugin_homework/.env.example",
    ]
    clean = True
    for name in files:
        path = ROOT / name
        original = path.read_bytes()
        normalized = (
            "\n".join(line.rstrip() for line in original.decode().splitlines()).rstrip()
            + "\n"
        ).encode()
        if original != normalized:
            if check:
                print("Whitespace differs: " + name)
                clean = False
            else:
                path.write_bytes(normalized)
    return clean


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--check",
        action="store_true",
        help="Report formatting differences without modifying files",
    )
    args = parser.parse_args()
    bin_dir = ".venv/Scripts" if os.name == "nt" else ".venv/bin"
    ruff_path = ROOT / bin_dir / ("ruff.exe" if os.name == "nt" else "ruff")
    ruff = str(ruff_path) if ruff_path.is_file() else shutil.which("ruff")
    prettier = ROOT / "web/node_modules/prettier/bin/prettier.cjs"
    taplo = ROOT / "web/node_modules/@taplo/cli/dist/cli.js"
    if not ruff or not prettier.is_file() or not taplo.is_file():
        print("Missing formatters. Run `make install-dev` first.", file=sys.stderr)
        return 1
    checks = [
        run(
            [
                "go",
                "tool",
                "goimports",
                "-local",
                "homework-station",
                "-l" if args.check else "-w",
                "cmd",
                "internal",
            ],
            ROOT / "backend",
            listed_files_fail=args.check,
        )
    ]
    checks.append(
        run(
            [
                "node",
                str(prettier),
                "--check" if args.check else "--write",
                "--ignore-unknown",
                ".",
            ]
        )
    )
    checks.append(
        run(
            [
                "node",
                str(taplo),
                "format",
                *(["--check"] if args.check else []),
                "pyproject.toml",
                ".taplo.toml",
            ]
        )
    )
    checks.append(run([ruff, "check", *([] if args.check else ["--fix"]), "."]))
    checks.append(run([ruff, "format", *(["--check"] if args.check else []), "."]))
    checks.append(go_module(args.check))
    checks.append(whitespace(args.check))
    return 0 if all(checks) else 1


if __name__ == "__main__":
    raise SystemExit(main())
