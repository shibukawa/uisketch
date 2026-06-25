export function createInitialRoot() {
  return {
    type: "browser",
    id: "screen-main",
    title: "UI Sketch Editor",
    address: "https://shibukawa.github.io/uisketch",
    children: [
      {
        type: "hstack",
        children: [
          { type: "label", label: "Screen title" },
          { type: "button", label: "Primary action", action: "action.primary" },
        ],
      },
      {
        type: "section",
        title: "Main content",
        children: [{ type: "table", id: "items-table", columns: ["Name", "Status"] }],
      },
    ],
  };
}

export function createRoot(type) {
  const titles = {
    browser: "Browser Sketch",
    window: "Window Sketch",
    mobile: "Mobile Sketch",
    dialog: "Dialog Sketch",
  };
  const root = {
    type,
    id: type === "browser" ? "screen-main" : `${type}-main`,
    children: [],
  };
  if (type !== "menu") root.title = titles[type] || "Sketch";
  if (type === "browser") root.address = "https://example.test";
  return root;
}

export function createDefaultNode(type, suffix) {
  switch (type) {
    case "button":
      return { type, id: `button-${suffix}`, label: "Button", action: "" };
    case "checkbox":
      return { type, id: `checkbox-${suffix}`, label: "Check box", action: "" };
    case "radio":
      return { type, id: `radio-${suffix}`, label: "Radio button", action: "" };
    case "toggle":
      return { type, id: `toggle-${suffix}`, label: "Toggle", action: "" };
    case "label":
      return { type, label: "Label" };
    case "input":
      return { type, id: `input-${suffix}`, label: "Input", hint: "text" };
    case "textarea":
      return { type, id: `textarea-${suffix}`, label: "Text area", hint: "multiline text" };
    case "select":
      return { type, id: `select-${suffix}`, label: "Combo box", options: ["Option 1", "Option 2"] };
    case "table":
      return { type, id: `table-${suffix}`, columns: ["Name", "Status"] };
    case "image":
      return { type, label: "Image" };
    case "list":
      return { type, label: "List" };
    case "section":
      return { type, title: "Section", children: [] };
    case "tabs":
      return {
        type,
        id: `tabs-${suffix}`,
        labels: [
          { text: "Tab A", selected: true },
          { text: "Tab B" },
        ],
        children: [{ type: "label", label: "Tab content" }],
      };
    case "grid":
      return { type, id: `grid-${suffix}`, columns: 2, children: [] };
    case "spacer":
      return { type };
    case "hstack":
    case "vstack":
      return { type, children: [] };
    default:
      return { type };
  }
}

export function displayName(node) {
  return node.id || node.title || node.label || node.action || "";
}

export function leafText(node) {
  if (node.type === "table") return (node.columns || []).join(" | ") || "Table";
  return node.label || node.title || node.id || node.type;
}
