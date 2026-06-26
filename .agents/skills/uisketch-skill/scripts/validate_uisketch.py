#!/usr/bin/env python3
"""Validate uisketch Markdown files and extracted source definitions."""

from __future__ import annotations

import argparse
import importlib.util
import json
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
VENDOR_PYTHON = SCRIPT_DIR / "vendor" / "python"
if VENDOR_PYTHON.exists():
    sys.path.insert(0, str(VENDOR_PYTHON))
sys.path.insert(0, str(SCRIPT_DIR))

import extract_uisketch  # noqa: E402


ROOT_TYPES = {"browser", "window", "mobile", "dialog", "menu"}
SUPPORTED_TYPES = {
    "browser",
    "window",
    "mobile",
    "dialog",
    "menu",
    "vstack",
    "hstack",
    "spacer",
    "grid",
    "splitter",
    "section",
    "tabs",
    "button",
    "toggle",
    "label",
    "note",
    "review",
    "image",
    "input",
    "textarea",
    "combobox",
    "checkbox",
    "radio",
    "slider",
    "table",
    "list",
    "tree",
    "calendar",
    "badge",
    "custom",
}
SCALAR_LABEL_TYPES = {
    "button",
    "toggle",
    "label",
    "note",
    "review",
    "badge",
    "input",
    "textarea",
    "combobox",
    "checkbox",
    "radio",
    "slider",
    "image",
    "custom",
    "table",
    "list",
    "tree",
    "calendar",
}


def optional_import(name: str):
    if importlib.util.find_spec(name) is None:
        return None
    try:
        return __import__(name)
    except Exception:
        return None


def finding(severity: str, message: str, index: int | None = None, line: int | None = None) -> dict[str, object]:
    out: dict[str, object] = {"severity": severity, "message": message}
    if index is not None:
        out["index"] = index
    if line is not None:
        out["line"] = line
    return out


def validate_frontmatter(path: str, frontmatter: dict[str, object]) -> list[dict[str, object]]:
    findings: list[dict[str, object]] = []
    if path.endswith(".uisketch.md"):
        if not frontmatter:
            findings.append(finding("error", "missing frontmatter"))
            return findings
        if frontmatter.get("type") != "uisketch":
            findings.append(finding("error", f"frontmatter type is {frontmatter.get('type')!r}, want 'uisketch'"))
        if not frontmatter.get("id"):
            findings.append(finding("error", "frontmatter id is required"))
    return findings


def load_yaml(source: str):
    yaml = optional_import("yaml")
    if yaml is None:
        return None, "YAML parser is not available; skipped YAML parse and schema validation"
    try:
        docs = list(yaml.safe_load_all(source))
    except Exception as exc:  # pragma: no cover - depends on optional library
        return None, f"malformed YAML: {exc}"
    if len(docs) != 1:
        return None, f"expected one YAML document, found {len(docs)}"
    return docs[0], None


def root_type(node: object) -> str | None:
    if isinstance(node, dict) and len(node) == 1:
        key = next(iter(node.keys()))
        return str(key)
    return None


def as_props(node: object) -> tuple[str | None, object]:
    if isinstance(node, str) and node in SUPPORTED_TYPES:
        return node, None
    if not isinstance(node, dict) or len(node) != 1:
        return None, None
    key = next(iter(node.keys()))
    return str(key), node[key]


def child_nodes(props: object) -> list[object]:
    if not isinstance(props, dict):
        return []
    out: list[object] = []
    for key in ("children", "buttons"):
        value = props.get(key)
        if isinstance(value, list):
            out.extend(value)
        elif isinstance(value, dict):
            out.append(value)
    child = props.get("child")
    if isinstance(child, dict):
        out.append(child)
    return out


def slot_values(value: object) -> tuple[int, int, int, bool]:
    if not isinstance(value, list):
        return 0, 0, 0, False
    total = 0
    stars = 0
    bad = False
    for item in value:
        if item in ("$", "*"):
            stars += 1
            continue
        if isinstance(item, int):
            total += item
        else:
            bad = True
    return len(value), total, stars, bad


def walk(node: object, callback):
    callback(node)
    typ, props = as_props(node)
    if typ and isinstance(props, dict):
        for child in child_nodes(props):
            walk(child, callback)


def semantic_validate(node: object, index: int, line: int) -> list[dict[str, object]]:
    findings: list[dict[str, object]] = []
    typ = root_type(node)
    if typ is None:
        return [finding("error", "layout root must be an object with exactly one component key", index, line)]
    if typ not in ROOT_TYPES:
        findings.append(finding("error", f"layout root {typ!r} is not a preferred root surface", index, line))

    ids: set[str] = set()

    def check(current: object):
        ctype, props = as_props(current)
        if ctype is None:
            findings.append(finding("error", "child node must be an object with exactly one component key", index))
            return
        if ctype not in SUPPORTED_TYPES:
            findings.append(finding("error", f"unsupported component type {ctype!r}", index))
        if props is None:
            return
        if not isinstance(props, dict):
            if ctype not in SCALAR_LABEL_TYPES:
                findings.append(finding("error", f"{ctype!r} does not support scalar label shorthand", index))
            return
        eid = props.get("id")
        if isinstance(eid, str) and eid:
            if eid in ids:
                findings.append(finding("error", f"duplicate element id {eid!r}", index))
            ids.add(eid)
        if "to" in props:
            findings.append(finding("warning", "legacy navigation field 'to' is deprecated; use 'anchor'", index))
        if "child" in props:
            findings.append(finding("warning", "legacy 'child' property should be rewritten as 'children'", index))
        if "prompt" in props and not isinstance(props["prompt"], str):
            findings.append(finding("error", "prompt must be a string", index))
        if ctype == "tabs":
            labels = props.get("labels")
            selected = 0
            if isinstance(labels, list):
                selected = sum(1 for item in labels if isinstance(item, list))
            if not labels:
                findings.append(finding("warning", "tabs should define labels", index))
            elif selected != 1:
                findings.append(finding("error", f"tabs labels must select exactly one label; found {selected}", index))
            children = props.get("children")
            if not children:
                findings.append(finding("error", "tabs children must define the active tab body", index))
            elif isinstance(children, list) and len(children) != 1:
                findings.append(finding("error", "tabs children must define exactly one active tab body", index))
        if "widths" in props:
            count, total, stars, bad = slot_values(props["widths"])
            child_count = len(child_nodes({"children": props.get("children", [])}))
            if ctype not in ("hstack", "tabs"):
                findings.append(finding("error", f"widths is only valid on hstack or tabs, not {ctype}", index))
            if ctype == "hstack" and count != child_count:
                findings.append(finding("warning", f"widths length {count} does not match children length {child_count}", index))
            if bad:
                findings.append(finding("error", "widths must contain integer percentages or '$'", index))
            if stars == 0 and ctype == "hstack" and total != 100:
                findings.append(finding("error", f"widths numeric percentages total {total}; must equal 100 when no '$' is present", index))
            if stars and total > 100:
                findings.append(finding("error", f"widths numeric percentages total {total}; must not exceed 100 with '$'", index))
        if "heights" in props:
            count, total, stars, bad = slot_values(props["heights"])
            child_count = len(child_nodes({"children": props.get("children", [])}))
            if ctype != "vstack":
                findings.append(finding("error", f"heights is only valid on vstack, not {ctype}", index))
            if ctype == "vstack" and count != child_count:
                findings.append(finding("warning", f"heights length {count} does not match children length {child_count}", index))
            if bad:
                findings.append(finding("error", "heights must contain integer percentages or '$'", index))
            if stars == 0 and ctype == "vstack" and total != 100:
                findings.append(finding("error", f"heights numeric percentages total {total}; must equal 100 when no '$' is present", index))
            if stars and total > 100:
                findings.append(finding("error", f"heights numeric percentages total {total}; must not exceed 100 with '$'", index))
        if "sizes" in props:
            count, total, stars, bad = slot_values(props["sizes"])
            child_count = len(child_nodes({"children": props.get("children", [])}))
            if ctype != "splitter":
                findings.append(finding("error", f"sizes is only valid on splitter, not {ctype}", index))
            if ctype == "splitter" and child_count != 2:
                findings.append(finding("warning", f"splitter should have exactly 2 children; found {child_count}", index))
            if ctype == "splitter" and count != 2:
                findings.append(finding("warning", f"splitter sizes length {count} should be 2", index))
            if bad:
                findings.append(finding("error", "sizes must contain integer percentages or '$'", index))
            if stars == 0 and ctype == "splitter" and total != 100:
                findings.append(finding("error", f"sizes numeric percentages total {total}; must equal 100 when no '$' is present", index))
            if stars and total > 100:
                findings.append(finding("error", f"sizes numeric percentages total {total}; must not exceed 100 with '$'", index))

    walk(node, check)
    return findings


def schema_validate(node: object, schema_path: str | None, index: int) -> list[dict[str, object]]:
    if not schema_path:
        return []
    schema = json.loads(Path(schema_path).read_text(encoding="utf-8"))
    fastjsonschema = optional_import("fastjsonschema")
    if fastjsonschema is None or not hasattr(fastjsonschema, "compile"):
        return [finding("error", "fastjsonschema is not available; bundled runtime is missing or broken", index)]
    try:
        validator = fastjsonschema.compile(schema)
        validator(node)
    except Exception as exc:  # pragma: no cover - depends on bundled library behavior
        return [finding("error", f"schema validation failed: {exc}", index)]
    return []


def validate_file(path: str, schema_path: str | None, include_yaml_fences: bool) -> list[dict[str, object]]:
    text = Path(path).read_text(encoding="utf-8")
    result = extract_uisketch.extract(text, include_yaml=include_yaml_fences)
    findings = validate_frontmatter(path, result.get("frontmatter", {}))
    definitions = result.get("definitions", [])
    if not definitions:
        findings.append(finding("error", "no uisketch source fences or generated source comments found"))
        return findings
    yaml_missing_warned = False
    for index, item in enumerate(definitions, start=1):
        if not isinstance(item, dict):
            continue
        source = str(item.get("source", ""))
        line = int(item.get("line", 0))
        if not source.strip():
            findings.append(finding("error", "empty uisketch source", index, line))
            continue
        if "--" in source and item.get("kind") == "generated-comment":
            findings.append(finding("error", "generated comment source cannot contain '--'", index, line))
        node, error = load_yaml(source)
        if error:
            severity = "warning" if error.startswith("YAML parser") else "error"
            if severity == "warning":
                if not yaml_missing_warned:
                    findings.append(finding(severity, error, index, line))
                    yaml_missing_warned = True
            else:
                findings.append(finding(severity, error, index, line))
            continue
        findings.extend(schema_validate(node, schema_path, index))
        findings.extend(semantic_validate(node, index, line))
    return findings


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("paths", nargs="+", help="Markdown or .uisketch.md files")
    parser.add_argument("--schema", help="JSON Schema path for optional schema validation")
    parser.add_argument("--include-yaml-fences", action="store_true", help="also validate bare yaml fences for migration")
    parser.add_argument("--json", action="store_true", help="emit JSON findings")
    args = parser.parse_args()

    all_findings: list[dict[str, object]] = []
    for path in args.paths:
        try:
            findings = validate_file(path, args.schema, args.include_yaml_fences)
        except Exception as exc:
            findings = [finding("error", str(exc))]
        for item in findings:
            item["path"] = path
        all_findings.extend(findings)

    if args.json:
        json.dump(all_findings, sys.stdout, ensure_ascii=False, indent=2)
        sys.stdout.write("\n")
    else:
        for item in all_findings:
            loc = item["path"]
            if "line" in item:
                loc = f"{loc}:{item['line']}"
            if "index" in item:
                loc = f"{loc} [source {item['index']}]"
            print(f"{item['severity']}: {loc}: {item['message']}")
        if not all_findings:
            print("ok")
    return 1 if any(item["severity"] == "error" for item in all_findings) else 0


if __name__ == "__main__":
    raise SystemExit(main())
