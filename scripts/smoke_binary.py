"""Verify a release executable runs with its embedded UI from an empty directory."""

import json
import os
import re
import shutil
import socket
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path


def main() -> None:
    source = Path(sys.argv[1]).resolve(strict=True)
    opener = urllib.request.build_opener(urllib.request.ProxyHandler({}))
    with tempfile.TemporaryDirectory(prefix="homework-binary-smoke-") as directory:
        root = Path(directory)
        binary = root / source.name
        shutil.copy2(source, binary)
        with socket.socket() as sock:
            sock.bind(("127.0.0.1", 0))
            port = sock.getsockname()[1]
        url = f"http://127.0.0.1:{port}"
        environment = {
            **os.environ,
            "PORT": str(port),
            "DATA_DIR": "./data/",
            "DB_DRIVER": "sqlite",
            "DB_DSN": "",
            "JWT_SECRET": "test-session-secret-32-characters-long",
            "GIN_MODE": "release",
        }
        with tempfile.TemporaryFile(mode="w+") as log:
            process = subprocess.Popen(
                [str(binary)], cwd=root, env=environment, stdout=log, stderr=log
            )
            try:
                for _ in range(200):
                    if process.poll() is not None:
                        raise RuntimeError("Release executable exited during startup")
                    try:
                        with opener.open(url + "/api/health", timeout=1) as response:
                            assert json.load(response) == {"ok": True}
                        break
                    except OSError:
                        time.sleep(0.1)
                else:
                    raise RuntimeError("Release executable did not become healthy")

                with opener.open(url + "/", timeout=5) as response:
                    html = response.read().decode()
                    assert "text/html" in response.headers["Content-Type"]
                assert '<div id="root">' in html
                with opener.open(url + "/admin/settings", timeout=5) as response:
                    assert response.read().decode() == html
                scripts = re.findall(r'<script\b[^>]*\bsrc="([^"]+)"', html)
                styles = re.findall(r'<link\b[^>]*\bhref="([^"]+\.css)"', html)
                assert scripts and styles, "Built frontend assets are missing"
                for asset in scripts + styles:
                    with opener.open(url + asset, timeout=5) as response:
                        assert response.status == 200 and response.read()
                        assert "text/html" not in response.headers["Content-Type"]
                with opener.open(url + "/api/public-settings", timeout=5) as response:
                    assert "site_title" in json.load(response)
                for path in ("/assets/missing.js", "/api/missing"):
                    try:
                        opener.open(url + path, timeout=5).close()
                    except urllib.error.HTTPError as error:
                        assert error.code == 404
                        error.close()
                    else:
                        raise AssertionError(f"Expected 404 for {path}")
                assert (root / "data/homework.db").is_file()
                assert set(path.name for path in root.iterdir()) == {binary.name, "data"}
                print("Single-file smoke test passed: UI, SPA routes, JS/CSS, API, data.")
            except BaseException:
                log.seek(0)
                print(log.read(), file=sys.stderr)
                raise
            finally:
                if process.poll() is None:
                    process.terminate()
                try:
                    process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait(timeout=10)


if __name__ == "__main__":
    main()
