#!/usr/bin/env python3
"""
Make Release Script for ptx-agent

Usage:
    python scripts/make-release.py [version]
    python scripts/make-release.py 0.1.0
    python scripts/make-release.py  # auto-detect from plugin.json

This script:
1. Validates version (or auto-detects from plugin.json)
2. Cross-compiles Go binary for all supported platforms
3. Creates dist/ptx-agent/ directory with binaries and plugin.json
4. Creates per-platform archives
5. Generates checksums
6. Creates release notes
"""

import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import zipfile
from datetime import datetime, timezone
from pathlib import Path


class Colors:
    RED = '\033[0;31m'
    GREEN = '\033[0;32m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'

    @classmethod
    def disable(cls):
        cls.RED = cls.GREEN = cls.YELLOW = cls.BLUE = cls.CYAN = cls.NC = ''


# Platforms to build (must match plugin.json os_support)
PLATFORMS = [
    ("linux", "amd64"),
    ("windows", "amd64"),
]

BINARY_NAME = "ptx-agent"
SRC_DIR = "src"


def print_header() -> None:
    print(f"{Colors.BLUE}========================================")
    print(f"  ptx-agent Release Maker")
    print(f"========================================{Colors.NC}")
    print()


def print_step(message: str) -> None:
    print(f"{Colors.GREEN}>> {message}{Colors.NC}")


def print_info(message: str) -> None:
    print(f"{Colors.CYAN}   {message}{Colors.NC}")


def print_error(message: str) -> None:
    print(f"{Colors.RED}ERROR: {message}{Colors.NC}")


def get_project_root() -> Path:
    script_dir = Path(__file__).parent.resolve()
    return script_dir.parent


def detect_version_from_plugin_json() -> str | None:
    """Auto-detect version from plugin.json"""
    project_root = get_project_root()
    plugin_json_path = project_root / "plugin.json"

    if not plugin_json_path.exists():
        return None

    try:
        content = json.loads(plugin_json_path.read_text(encoding='utf-8'))
        return content.get('version')
    except (json.JSONDecodeError, KeyError):
        return None


def binary_name_for_platform(target_os: str) -> str:
    """Return binary filename with platform-appropriate extension"""
    if target_os == "windows":
        return f"{BINARY_NAME}.exe"
    return BINARY_NAME


def calculate_sha256(file_path: Path) -> str:
    sha256_hash = hashlib.sha256()
    with open(file_path, 'rb') as f:
        for chunk in iter(lambda: f.read(4096), b''):
            sha256_hash.update(chunk)
    return sha256_hash.hexdigest()


def cross_compile(project_root: Path, target_os: str, target_arch: str, output_path: Path) -> bool:
    """Cross-compile Go binary for given platform"""
    env = os.environ.copy()
    env["GOOS"] = target_os
    env["GOARCH"] = target_arch
    env["CGO_ENABLED"] = "0"

    cmd = ["go", "build", "-o", str(output_path), f"./cmd/{BINARY_NAME}/"]
    print_info(f"GOOS={target_os} GOARCH={target_arch} go build -o {output_path.name}")

    result = subprocess.run(cmd, cwd=project_root / SRC_DIR, env=env, capture_output=True, text=True)
    if result.returncode != 0:
        print_error(f"Build failed for {target_os}/{target_arch}")
        if result.stderr:
            print(result.stderr)
        return False

    return True


def create_release(version: str) -> bool:
    """Create release for ptx-agent across all platforms"""
    project_root = get_project_root()
    plugin_json_src = project_root / "plugin.json"

    if not plugin_json_src.exists():
        print_error(f"plugin.json not found in {project_root}")
        return False

    # Create dist directory
    dist_dir = project_root / "dist"
    plugin_dist = dist_dir / BINARY_NAME
    if plugin_dist.exists():
        shutil.rmtree(plugin_dist)
    plugin_dist.mkdir(parents=True, exist_ok=True)

    # Copy plugin.json
    shutil.copy2(plugin_json_src, plugin_dist / "plugin.json")
    print_info("Copied plugin.json")

    # Cross-compile for each platform
    print_step("Cross-compiling for all platforms...")
    built_binaries = []
    for target_os, target_arch in PLATFORMS:
        bin_name = binary_name_for_platform(target_os)
        output_path = plugin_dist / bin_name

        if not cross_compile(project_root, target_os, target_arch, output_path):
            return False

        # Set executable permission on Unix binaries
        if target_os != "windows":
            output_path.chmod(0o755)

        built_binaries.append((target_os, target_arch, bin_name))
        print_info(f"Built {bin_name} ({target_os}/{target_arch})")

    # Create per-platform archives
    print_step("Creating archives...")
    archive_paths = []
    for target_os, target_arch, bin_name in built_binaries:
        archive_base = f"{BINARY_NAME}-{version}-{target_os}-{target_arch}"
        bin_path = plugin_dist / bin_name

        if target_os == "windows":
            # ZIP for Windows
            zip_path = dist_dir / f"{archive_base}.zip"
            with zipfile.ZipFile(zip_path, 'w', zipfile.ZIP_DEFLATED) as zipf:
                zipf.write(bin_path, f"{archive_base}/{bin_name}")
                zipf.write(plugin_dist / "plugin.json", f"{archive_base}/plugin.json")
            print_info(f"Created: {zip_path.name}")
            archive_paths.append(zip_path)
        else:
            # tar.gz for Linux
            tar_path = dist_dir / f"{archive_base}.tar.gz"
            with tarfile.open(tar_path, "w:gz") as tar:
                tar.add(bin_path, arcname=f"{archive_base}/{bin_name}")
                tar.add(plugin_dist / "plugin.json", arcname=f"{archive_base}/plugin.json")
            print_info(f"Created: {tar_path.name}")
            archive_paths.append(tar_path)

    # Generate checksums
    print_step("Generating checksums...")
    checksums_path = dist_dir / f"checksums_{version}.txt"
    with open(checksums_path, 'w') as f:
        for archive_path in archive_paths:
            checksum = calculate_sha256(archive_path)
            f.write(f"{checksum}  {archive_path.name}\n")
    print_info(f"Created: {checksums_path.name}")

    # Create release notes
    print_step("Creating release notes...")
    create_release_notes(dist_dir, version)

    return True


def create_release_notes(dist_dir: Path, version: str) -> None:
    build_date = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%S UTC")
    platforms = ", ".join(f"{os}/{arch}" for os, arch in PLATFORMS)

    notes = f"""# ptx-agent v{version}

Remote device control agent for Portunix — file browsing, accessibility, keyboard, browser automation.

## Features

- File browsing on target machine
- VS Code control via accessibility APIs
- Application control (keyboard, windows)
- MCP server for AI assistant integration
- Browser automation via Playwright

## Platforms

{platforms}

## Installation

```bash
portunix plugin install plugin.json
portunix plugin enable ptx-agent
```

## Verification

Verify download using SHA256 checksums in `checksums_{version}.txt`.

---

**Build Date:** {build_date}
**Repository:** https://github.com/CassandraGargoyle/portunix-agent
"""

    notes_path = dist_dir / f"RELEASE_NOTES_{version}.md"
    notes_path.write_text(notes)
    print_info(f"Created: {notes_path.name}")


def show_summary(version: str) -> None:
    project_root = get_project_root()
    dist_dir = project_root / "dist"

    print()
    print(f"{Colors.GREEN}========================================")
    print(f"  Release {version} ready!")
    print(f"========================================{Colors.NC}")
    print()
    print("Generated files:")
    for f in sorted(dist_dir.iterdir()):
        if f.is_file():
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"  {f.name} ({size_mb:.1f} MB)")
    if (dist_dir / BINARY_NAME).is_dir():
        print(f"  {BINARY_NAME}/")
        for f in sorted((dist_dir / BINARY_NAME).iterdir()):
            size_mb = f.stat().st_size / (1024 * 1024)
            print(f"    {f.name} ({size_mb:.1f} MB)")
    print()
    print("Next steps:")
    print(f"  1. Test: portunix plugin install dist/{BINARY_NAME}/plugin.json")
    print(f"  2. Create GitHub release with files from dist/")
    print()


def show_usage() -> None:
    print("Usage: python scripts/make-release.py [version]")
    print()
    print("Examples:")
    print("  python scripts/make-release.py 0.1.0")
    print("  python scripts/make-release.py  # auto-detect from plugin.json")


def main() -> int:
    if not sys.stdout.isatty():
        Colors.disable()

    print_header()

    if len(sys.argv) >= 2 and sys.argv[1] in ('-h', '--help'):
        show_usage()
        return 0

    # Get version from argument or auto-detect
    if len(sys.argv) >= 2:
        version = sys.argv[1]
    else:
        print_info("Auto-detecting version from plugin.json...")
        version = detect_version_from_plugin_json()
        if not version:
            print_error("Could not detect version from plugin.json")
            print_info("Please provide version: python scripts/make-release.py 0.1.0")
            return 1
        print_info(f"Detected version: {version}")

    # Validate version format
    if not re.match(r'^\d+\.\d+\.\d+(-\w+)?$', version):
        print_error("Invalid version format. Use: X.Y.Z or X.Y.Z-suffix")
        return 1

    print_info(f"Creating release for {BINARY_NAME} v{version}")
    print()

    if not create_release(version):
        return 1

    show_summary(version)
    return 0


if __name__ == "__main__":
    sys.exit(main())
