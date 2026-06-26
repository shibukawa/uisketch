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
  if (type === "window") root.menu = ["File", "Edit", "View"];
  if (type === "mobile") root.menu = ["Home", "Search", "Settings"];
  if (type === "dialog") root.buttons = [
    { type: "button", label: "Cancel" },
    { type: "button", label: "OK" },
  ];
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
    case "combobox":
      return { type, id: `combobox-${suffix}`, label: "Combo box", hint: "select", options: ["Option 1", "Option 2"] };
    case "slider":
      return { type, id: `slider-${suffix}`, label: "Slider" };
    case "table":
      return { type, id: `table-${suffix}`, columns: ["Name", "Status"] };
    case "image":
      return { type, label: "Image" };
    case "custom":
      return { type, name: "custom-widget", purpose: "Project-defined component" };
    case "list":
      return { type, label: "List" };
    case "tree":
      return { type, label: "Tree" };
    case "calendar":
      return { type, label: "Calendar" };
    case "badge":
      return { type, label: "Badge" };
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
    case "splitter":
      return {
        type,
        id: `splitter-${suffix}`,
        orientation: "horizontal",
        sizes: [25, 75],
        children: [
          { type: "section", title: "Primary", children: [] },
          { type: "section", title: "Secondary", children: [] },
        ],
      };
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
  return node.label || node.title || node.name || node.purpose || node.id || node.type;
}
