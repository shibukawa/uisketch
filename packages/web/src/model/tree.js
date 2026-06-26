export function getNode(root, path) {
  let node = root;
  let collection = "children";
  for (const segment of path) {
    if (segment === "buttons") {
      collection = "buttons";
      continue;
    }
    node = node?.[collection]?.[segment];
    collection = "children";
  }
  return node || null;
}

export function walk(node, path, visit) {
  visit(node, path);
  (node.children || []).forEach((child, index) => walk(child, [...path, index], visit));
  (node.buttons || []).forEach((button, index) => walk(button, [...path, "buttons", index], visit));
}

export function samePath(a, b) {
  return a.length === b.length && a.every((value, index) => value === b[index]);
}

export function isDescendantPath(path, maybeAncestor) {
  return maybeAncestor.length <= path.length && maybeAncestor.every((value, index) => value === path[index]);
}
