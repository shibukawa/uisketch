import { parseLayoutYaml, serializeYaml } from "./yaml.js";

export function serializeSourceDocument(root, name = "Untitled Sketch") {
  const id = root.id || slugify(name) || "screen-main";
  const title = root.title || name || "Untitled Sketch";
  return `---\nid: ${quoteYaml(id)}\ntype: uisketch\ntitle: ${quoteYaml(title)}\n---\n\n# ${title}\n\n## Layout\n\n\`\`\`uisketch\n${serializeYaml(root)}\`\`\`\n`;
}

export function parseSourceDocument(source) {
  const fenced = source.match(/```uisketch(?::[A-Za-z]+)?\n([\s\S]*?)\n```/);
  const layoutSource = fenced ? fenced[1] : source;
  return parseLayoutYaml(layoutSource);
}

export function slugify(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9_-]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function quoteYaml(value) {
  return JSON.stringify(String(value || ""));
}
