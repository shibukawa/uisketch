# uisketch-skill

This skill helps Codex create, inspect, extract, and validate uisketch Markdown files.

## Usage

Extract every renderable uisketch definition:

```bash
./scripts/extract_uisketch.py path/to/file.uisketch.md
```

Validate a sketch with the bundled schema:

```bash
./scripts/validate_uisketch.py path/to/file.uisketch.md --schema references/uisketch-layout.schema.json
```

Validate older examples that still use a bare `yaml` fence under a layout heading:

```bash
./scripts/validate_uisketch.py path/to/file.uisketch.md --include-yaml-fences --schema references/uisketch-layout.schema.json
```

## Runtime Dependencies

The validation script first tries bundled runtime libraries under `scripts/vendor/python`.

- YAML parsing uses the bundled `yaml` module from `pyyaml-pure`.
- JSON Schema validation uses bundled `fastjsonschema`.

The bundled uisketch schema declares JSON Schema draft 2020-12, but it intentionally uses a small keyword subset that `fastjsonschema` can compile.

## Bundled Libraries

| Library | Source | Author | License | Used by runtime |
| --- | --- | --- | --- | --- |
| `pyyaml-pure` | `assets/vendor/source/pyyaml-pure` | Ali Fadel | MIT | Yes, via `scripts/vendor/python/yaml` |
| `fastjsonschema` | `assets/vendor/source/fastjsonschema` | Michal Horejsek | BSD-3-Clause | Yes, via `scripts/vendor/python/fastjsonschema` |

License files are preserved in the bundled source trees:

- `assets/vendor/source/pyyaml-pure/LICENSE`
- `assets/vendor/source/fastjsonschema/LICENSE`

## Update Notes

When updating bundled libraries, replace both the original source tree under `assets/vendor/source` and any runtime package copied under `scripts/vendor/python`, then rerun validation against representative uisketch fixtures.
