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
  const props = Object.entries(node).filter(([key, value]) => key !== "type" && key !== "children" && valueIsPresent(value));

  for (const [key, value] of props) {
    if (key === "labels" && Array.isArray(value)) {
      lines.push(`${propPad}${key}:`);
      for (const item of value) {
        const text = item?.text || "";
        lines.push(`${propPad}  - ${item?.selected ? `[${escapeYaml(text)}]` : escapeYaml(text)}`);
      }
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

    if (key === "children") {
      node.children = [];
      while (index < lines.length) {
        index = firstContentLine(lines, index);
        if (index >= lines.length || countIndent(lines[index]) < propIndent + 2) break;
        const child = parseNode(lines, index, propIndent + 2, true);
        if (!child.node) break;
        node.children.push(child.node);
        index = child.index;
      }
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

function parseTabLabel(value) {
  if (typeof value === "string" && value.startsWith("[") && value.endsWith("]")) {
    return { text: value.slice(1, -1), selected: true };
  }
  return { text: String(value) };
}

function parseScalar(value) {
  if (value === '""') return "";
  if (/^-?\d+$/.test(value)) return Number(value);
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
