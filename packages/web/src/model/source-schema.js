import { canHaveChildren, fieldsFor, paletteGroups, rootSurfaceTypes } from "./component-catalog.js";

export const componentTypes = [...new Set([...rootSurfaceTypes, ...paletteGroups.flatMap((group) => group.items.map((item) => item.type))])].sort();

export const commonProperties = ["children"];
export const structuralProperties = ["columns", "labels", "widths", "heights"];
export const propertyTypes = [...new Set([...commonProperties, ...structuralProperties, ...componentTypes.flatMap((type) => fieldsFor({ type }))])].sort();
export const sourceKeywords = [...new Set([...componentTypes, ...propertyTypes])].sort();

export function completionsForContext(source, offset) {
  const component = nearestComponentBefore(source, offset);
  if (isNodeHeaderPosition(source, offset) && (!component || currentLinePrefix(source, offset).includes("-"))) return componentTypes;
  if (!component) return sourceKeywords;
  return propertiesForComponent(component);
}

export function sourceSchemaFindings(source) {
  const findings = [];
  const stack = [];
  source.split("\n").forEach((line, index) => {
    const match = line.match(/^(\s*)-?\s*([A-Za-z][A-Za-z0-9_-]*):/);
    if (!match) return;
    const indent = match[1].length;
    const key = match[2];
    while (stack.length && stack.at(-1).indent >= indent) stack.pop();
    if (componentTypes.includes(key)) {
      stack.push({ indent, type: key });
      return;
    }
    if (!propertyTypes.includes(key)) {
      findings.push({ severity: "Warning", message: `Line ${index + 1}: unknown UI Layout DSL keyword "${key}"` });
      return;
    }
    const owner = stack.at(-1)?.type;
    if (owner && !propertiesForComponent(owner).includes(key)) {
      findings.push({ severity: "Warning", message: `Line ${index + 1}: "${key}" is not a usual property for ${owner}` });
    }
  });
  return findings;
}

export function propertiesForComponent(type) {
  const base = new Set(fieldsFor({ type }));
  if (canHaveChildren({ type })) base.add("children");
  if (type === "grid") base.add("columns");
  if (type === "tabs") base.add("labels");
  if (type === "hstack") base.add("widths");
  if (type === "vstack") base.add("heights");
  return [...base].sort();
}

function isNodeHeaderPosition(source, offset) {
  return /^\s*-?\s*$/.test(currentLinePrefix(source, offset));
}

function currentLinePrefix(source, offset) {
  const lineStart = source.lastIndexOf("\n", Math.max(0, offset - 1)) + 1;
  return source.slice(lineStart, offset);
}

function nearestComponentBefore(source, offset) {
  const before = source.slice(0, offset).split("\n");
  const stack = [];
  for (const line of before) {
    const match = line.match(/^(\s*)-?\s*([A-Za-z][A-Za-z0-9_-]*):/);
    if (!match) continue;
    const indent = match[1].length;
    const key = match[2];
    while (stack.length && stack.at(-1).indent >= indent) stack.pop();
    if (componentTypes.includes(key)) stack.push({ indent, type: key });
  }
  return stack.at(-1)?.type || "";
}
