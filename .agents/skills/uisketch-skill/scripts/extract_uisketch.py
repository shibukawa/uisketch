#!/usr/bin/env python3
"""Extract uisketch source definitions from Markdown."""

from __future__ import annotations

import argparse
import json
import re
import sys
from pathlib import Path


FENCE_RE = re.compile(r"^```([^`]*)\s*$")
COMMENT_RE = re.compile(r"^<!--\s+uisketch:source\b(.*)$")
ATTR_RE = re.compile(r'([A-Za-z_][A-Za-z0-9_-]*)="([^"]*)"')
IMAGE_RE = re.compile(r"^!\[uisketch:([A-Za-z0-9_.:-]+)\]\(([^)]+)\)$")


def normalize_info(info: str, include_yaml: bool = False) -> tuple[str, bool]:
    info = info.strip()
    if info in ("uisketch", "uisketch:svg"):
        return "svg", True
    if info in ("uisketch:txt", "uisketch:text", "uisketch:ascii"):
        return "txt", True
    if include_yaml and info == "yaml":
        return "yaml", True
    return "", False


def split_frontmatter(text: str) -> tuple[str | None, str, int]:
    text = text.replace("\r\n", "\n")
    if not text.startswith("---\n"):
        return None, text, 1
    rest = text[4:]
    marker = rest.find("\n---")
    if marker < 0:
        return None, text, 1
    frontmatter = rest[:marker]
    body = rest[marker + 4 :]
    body_start = frontmatter.count("\n") + 4
    if body.startswith("\n"):
        body = body[1:]
    else:
        body_start -= 1
    return frontmatter, body, body_start


def parse_frontmatter_fields(frontmatter: str | None) -> dict[str, object]:
    if frontmatter is None:
        return {}
    out: dict[str, object] = {"raw": frontmatter}
    section = ""
    for raw in frontmatter.splitlines():
        if not raw.strip():
            continue
        indent = len(raw) - len(raw.lstrip(" "))
        stripped = raw.strip()
        if ":" not in stripped:
            continue
        key, value = stripped.split(":", 1)
        key = key.strip()
        value = value.strip().strip("\"'")
        if indent == 0:
            section = "" if value else key
        if indent == 0 and value:
            out[key] = value
        elif section == "screen" and key == "id" and value:
            screen = out.setdefault("screen", {})
            if isinstance(screen, dict):
                screen["id"] = value
    return out


def find_fence_end(lines: list[str], start: int) -> int:
    for i in range(start, len(lines)):
        if lines[i].strip() == "```":
            return i
    return -1


def parse_comment(lines: list[str], start: int, base_line: int) -> tuple[dict[str, object], int]:
    attrs = dict(ATTR_RE.findall(lines[start]))
    for i in range(start + 1, len(lines)):
        if lines[i].strip() == "-->":
            raise ValueError(f"uisketch source comment at line {base_line + start} has no source fence")
        match = FENCE_RE.match(lines[i])
        if not match:
            continue
        fmt, ok = normalize_info(match.group(1))
        if not ok:
            continue
        end = find_fence_end(lines, i + 1)
        if end < 0:
            raise ValueError(f"unterminated uisketch source fence at line {base_line + i}")
        if end + 1 >= len(lines) or lines[end + 1].strip() != "-->":
            raise ValueError(f"uisketch source comment at line {base_line + start} is not closed after source fence")
        source = "\n".join(lines[i + 1 : end])
        return (
            {
                "kind": "generated-comment",
                "id": attrs.get("id", ""),
                "format": attrs.get("format", fmt),
                "line": base_line + i + 1,
                "source": source,
            },
            end + 2,
        )
    raise ValueError(f"unterminated uisketch source comment at line {base_line + start}")


def extract(text: str, include_yaml: bool = False) -> dict[str, object]:
    frontmatter, body, body_start = split_frontmatter(text)
    lines = body.split("\n")
    definitions: list[dict[str, object]] = []
    i = 0
    while i < len(lines):
        if COMMENT_RE.match(lines[i].strip()):
            block, i = parse_comment(lines, i, body_start)
            definitions.append(block)
            continue
        match = FENCE_RE.match(lines[i])
        if match:
            fmt, ok = normalize_info(match.group(1), include_yaml=include_yaml)
            end = find_fence_end(lines, i + 1)
            if end < 0:
                raise ValueError(f"unterminated fence at line {body_start + i}")
            if ok:
                definitions.append(
                    {
                        "kind": "fence",
                        "id": "",
                        "format": fmt,
                        "line": body_start + i + 1,
                        "source": "\n".join(lines[i + 1 : end]),
                    }
                )
            i = end + 1
            continue
        i += 1
    return {
        "frontmatter": parse_frontmatter_fields(frontmatter),
        "definitions": definitions,
    }


def yaml_quote(value: object) -> str:
    return json.dumps("" if value is None else str(value), ensure_ascii=False)


def write_yaml_summary(result: dict[str, object]) -> str:
    lines: list[str] = []
    frontmatter = result["frontmatter"]
    lines.append("frontmatter:")
    if isinstance(frontmatter, dict) and frontmatter:
        for key in ("id", "type", "title", "status"):
            if key in frontmatter:
                lines.append(f"  {key}: {yaml_quote(frontmatter[key])}")
        screen = frontmatter.get("screen")
        if isinstance(screen, dict) and "id" in screen:
            lines.append("  screen:")
            lines.append(f"    id: {yaml_quote(screen['id'])}")
    else:
        lines.append("  {}")
    lines.append("definitions:")
    for idx, item in enumerate(result["definitions"], start=1):
        if not isinstance(item, dict):
            continue
        lines.append(f"  - index: {idx}")
        lines.append(f"    kind: {yaml_quote(item.get('kind', ''))}")
        lines.append(f"    format: {yaml_quote(item.get('format', ''))}")
        lines.append(f"    line: {item.get('line', 0)}")
        if item.get("id"):
            lines.append(f"    id: {yaml_quote(item.get('id'))}")
        lines.append("    source: |")
        source = str(item.get("source", ""))
        for source_line in source.splitlines() or [""]:
            lines.append(f"      {source_line}")
    return "\n".join(lines) + "\n"


def write_body_only(result: dict[str, object]) -> str:
    docs = []
    for item in result["definitions"]:
        if isinstance(item, dict):
            docs.append(str(item.get("source", "")).rstrip())
    return "\n---\n".join(docs) + ("\n" if docs else "")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("path", help="Markdown or .uisketch.md file")
    parser.add_argument("--json", action="store_true", help="emit JSON instead of YAML summary")
    parser.add_argument("--body-only", action="store_true", help="emit only raw YAML source bodies as multi-doc YAML")
    parser.add_argument("--include-yaml-fences", action="store_true", help="also extract bare yaml fences for migration")
    args = parser.parse_args()

    text = Path(args.path).read_text(encoding="utf-8")
    try:
        result = extract(text, include_yaml=args.include_yaml_fences)
    except ValueError as exc:
        print(f"error: {exc}", file=sys.stderr)
        return 2

    if args.body_only:
        sys.stdout.write(write_body_only(result))
    elif args.json:
        json.dump(result, sys.stdout, ensure_ascii=False, indent=2)
        sys.stdout.write("\n")
    else:
        sys.stdout.write(write_yaml_summary(result))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
