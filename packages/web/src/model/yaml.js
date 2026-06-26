export function serializeYaml(root) {
  return writeNode(root, 0, false).join("\n") + "\n";
}

export function parseLayoutYaml(source) {
  const lines = source.replace(/\r\n?/g, "\n").split("\n");
  const parsed = parseNode(lines, firstContentLine(lines, 0), 0, false);
  if (!parsed.node) throw new Error("No layout root found");
  return parsed.node;
}

function writeNode(node, indent, listItem) {
  const pad = " ".repeat(indent);
  const lines = [listItem ? `${pad}- ${node.type}:` : `${pad}${node.type}:`];
  const propIndent = indent + (listItem ? 4 : 2);
  const propPad = " ".repeat(propIndent);
  const props = Object.entries(node).filter(([key, value]) => key !== "type" && key !== "children" && key !== "buttons" && (key === "data" ? value !== "" && value !== undefined : valueIsPresent(value)));

  for (const [key, value] of props) {
    if (key === "labels" && Array.isArray(value)) {
      lines.push(`${propPad}${key}:`);
      for (const item of value) {
        const text = item?.text || "";
        lines.push(`${propPad}  - ${item?.selected ? `[${escapeYaml(text)}]` : escapeYaml(text)}`);
      }
      continue;
    }
    if (key === "data") {
      writeYamlProperty(lines, key, value, propIndent);
      continue;
    }
    if (Array.isArray(value)) {
      lines.push(`${propPad}${key}:`);
      for (const item of value) lines.push(`${propPad}  - ${escapeYaml(String(item))}`);
    } else {
      lines.push(`${propPad}${key}: ${escapeYaml(String(value))}`);
    }
  }

  if (node.children?.length) {
    lines.push(`${propPad}children:`);
    for (const child of node.children) {
      lines.push(...writeNode(child, propIndent + 2, true));
    }
  }
  if (node.buttons?.length) {
    lines.push(`${propPad}buttons:`);
    for (const button of node.buttons) {
      lines.push(...writeNode(button, propIndent + 2, true));
    }
  }
  return lines;
}

function valueIsPresent(value) {
  if (Array.isArray(value)) return value.length > 0;
  return value !== "" && value != null;
}

function escapeYaml(value) {
  if (!value) return '""';
  if (/[:#\[\]{},&*?|\-<>=!%@`]|^\s|\s$/.test(value)) return JSON.stringify(value);
  return value;
}

function firstContentLine(lines, index) {
  let i = index;
  while (i < lines.length && !lines[i].trim()) i += 1;
  return i;
}

function parseNode(lines, index, indent, listItem) {
  index = firstContentLine(lines, index);
  if (index >= lines.length) return { node: null, index };
  const line = lines[index];
  if (countIndent(line) !== indent) return { node: null, index };
  const header = listItem ? line.match(/^\s*-\s+([A-Za-z0-9_-]+):\s*$/) : line.match(/^\s*([A-Za-z0-9_-]+):\s*$/);
  if (!header) throw new Error(`Expected layout node at line ${index + 1}`);
  const node = { type: header[1] };
  index += 1;
  const propIndent = indent + (listItem ? 4 : 2);

  while (index < lines.length) {
    index = firstContentLine(lines, index);
    if (index >= lines.length) break;
    const currentIndent = countIndent(lines[index]);
    if (currentIndent < propIndent) break;
    if (currentIndent !== propIndent) throw new Error(`Unexpected indentation at line ${index + 1}`);
    const prop = lines[index].match(/^\s*([A-Za-z0-9_-]+):(.*)$/);
    if (!prop) throw new Error(`Expected property at line ${index + 1}`);
    const key = prop[1];
    const rest = prop[2].trim();
    index += 1;

    if (key === "children" || key === "buttons") {
      const target = [];
      while (index < lines.length) {
        index = firstContentLine(lines, index);
        if (index >= lines.length || countIndent(lines[index]) < propIndent + 2) break;
        const child = parseNode(lines, index, propIndent + 2, true);
        if (!child.node) break;
        target.push(child.node);
        index = child.index;
      }
      node[key] = target;
      continue;
    }

    if (key === "data") {
      if (rest !== "") {
        setNodeValue(node, key, parseScalar(rest));
        continue;
      }
      const parsed = parseYamlValueBlock(lines, index, propIndent + 2);
      setNodeValue(node, key, parsed.value);
      index = parsed.index;
      continue;
    }

    if (rest !== "") {
      setNodeValue(node, key, parseScalar(rest));
      continue;
    }

    const items = [];
    while (index < lines.length) {
      index = firstContentLine(lines, index);
      if (index >= lines.length || countIndent(lines[index]) < propIndent + 2) break;
      const item = lines[index].match(/^\s*-\s+(.*)$/);
      if (!item) break;
      items.push(parseScalar(item[1].trim()));
      index += 1;
    }
    setNodeValue(node, key, key === "labels" ? items.map(parseTabLabel) : items);
  }
  return { node, index };
}

function setNodeValue(node, key, value) {
  if (key === "columns" && typeof value === "number") node.gridColumns = value;
  node[key] = value;
}

export function serializeYamlValue(value) {
  if (!isStructuredYamlValue(value)) return escapeYaml(String(value ?? ""));
  const lines = [];
  writeNestedYamlValue(lines, value, 0);
  return lines.join("\n");
}

export function parseYamlValue(source) {
  const lines = String(source || "").replace(/\r\n?/g, "\n").split("\n");
  const index = firstContentLine(lines, 0);
  if (index >= lines.length) return "";
  if (lines[index].trim().startsWith("- ")) return parseYamlValueBlock(lines, index, countIndent(lines[index])).value;
  if (/^[A-Za-z0-9_-]+:/.test(lines[index].trim())) return parseYamlValueBlock(lines, index, countIndent(lines[index])).value;
  if (lines.length - index === 1) return parseScalar(lines[index].trim());
  return lines.slice(index).join("\n").trimEnd();
}

function writeYamlProperty(lines, key, value, indent) {
  const pad = " ".repeat(indent);
  if (!isStructuredYamlValue(value)) {
    lines.push(`${pad}${key}: ${escapeYaml(String(value ?? ""))}`);
    return;
  }
  lines.push(`${pad}${key}:`);
  writeNestedYamlValue(lines, value, indent + 2);
}

function writeNestedYamlValue(lines, value, indent) {
  const pad = " ".repeat(indent);
  if (Array.isArray(value)) {
    if (!value.length) {
      lines.push(`${pad}[]`);
      return;
    }
    for (const item of value) {
      if (isPlainObject(item) || Array.isArray(item)) {
        lines.push(`${pad}-`);
        writeNestedYamlValue(lines, item, indent + 2);
      } else {
        lines.push(`${pad}- ${escapeYamlScalar(item)}`);
      }
    }
    return;
  }
  if (isPlainObject(value)) {
    const entries = Object.entries(value);
    if (!entries.length) {
      lines.push(`${pad}{}`);
      return;
    }
    for (const [key, item] of entries) {
      if (isStructuredYamlValue(item)) {
        lines.push(`${pad}${key}:`);
        writeNestedYamlValue(lines, item, indent + 2);
      } else {
        lines.push(`${pad}${key}: ${escapeYamlScalar(item)}`);
      }
    }
    return;
  }
  lines.push(`${pad}${escapeYamlScalar(value)}`);
}

function parseYamlValueBlock(lines, index, indent) {
  index = firstContentLine(lines, index);
  if (index >= lines.length || countIndent(lines[index]) < indent) return { value: {}, index };
  if (lines[index].trim() === "[]") return { value: [], index: index + 1 };
  if (lines[index].trim() === "{}") return { value: {}, index: index + 1 };
  if (lines[index].trim().startsWith("-")) return parseYamlSequence(lines, index, indent);
  return parseYamlMapping(lines, index, indent);
}

function parseYamlMapping(lines, index, indent) {
  const out = {};
  while (index < lines.length) {
    index = firstContentLine(lines, index);
    if (index >= lines.length || countIndent(lines[index]) < indent) break;
    if (countIndent(lines[index]) !== indent) throw new Error(`Unexpected data indentation at line ${index + 1}`);
    const prop = lines[index].match(/^\s*([A-Za-z0-9_-]+):(.*)$/);
    if (!prop) break;
    const key = prop[1];
    const rest = prop[2].trim();
    index += 1;
    if (rest !== "") {
      out[key] = parseScalar(rest);
      continue;
    }
    const parsed = parseYamlValueBlock(lines, index, indent + 2);
    out[key] = parsed.value;
    index = parsed.index;
  }
  return { value: out, index };
}

function parseYamlSequence(lines, index, indent) {
  const out = [];
  while (index < lines.length) {
    index = firstContentLine(lines, index);
    if (index >= lines.length || countIndent(lines[index]) < indent) break;
    if (countIndent(lines[index]) !== indent) throw new Error(`Unexpected data indentation at line ${index + 1}`);
    const item = lines[index].match(/^\s*-\s*(.*)$/);
    if (!item) break;
    const rest = item[1].trim();
    index += 1;
    if (rest !== "") {
      out.push(parseScalar(rest));
      continue;
    }
    const parsed = parseYamlValueBlock(lines, index, indent + 2);
    out.push(parsed.value);
    index = parsed.index;
  }
  return { value: out, index };
}

function isStructuredYamlValue(value) {
  return Array.isArray(value) || isPlainObject(value);
}

function isPlainObject(value) {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}

function escapeYamlScalar(value) {
  if (value === null) return "null";
  if (typeof value === "boolean") return value ? "true" : "false";
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  return escapeYaml(String(value ?? ""));
}

function parseTabLabel(value) {
  if (typeof value === "string" && value.startsWith("[") && value.endsWith("]")) {
    return { text: value.slice(1, -1), selected: true };
  }
  return { text: String(value) };
}

function parseScalar(value) {
  if (value === '""') return "";
  if (value === "null" || value === "~") return null;
  if (value === "true") return true;
  if (value === "false") return false;
  if (value.startsWith("[") && value.endsWith("]")) {
    return value
      .slice(1, -1)
      .split(",")
      .map((item) => parseScalar(item.trim()))
      .filter((item) => item !== "");
  }
  if (/^-?\d+$/.test(value)) return Number(value);
  if (/^-?(?:\d+\.\d+|\d+\.)$/.test(value)) return Number(value);
  if ((value.startsWith('"') && value.endsWith('"')) || (value.startsWith("'") && value.endsWith("'"))) {
    try {
      return JSON.parse(value);
    } catch {
      return value.slice(1, -1);
    }
  }
  return value;
}

function countIndent(line) {
  return line.match(/^ */)[0].length;
}
