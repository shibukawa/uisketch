export const rootSurfaceTypes = ["browser", "window", "mobile", "dialog", "menu"];

export const paletteGroups = [
  {
    id: "layout",
    label: "Layout",
    items: [
      { type: "vstack", label: "V Stack", icon: "V" },
      { type: "hstack", label: "H Stack", icon: "H" },
      { type: "grid", label: "Grid", icon: "Grid" },
      { type: "splitter", label: "Splitter", icon: "Split" },
      { type: "spacer", label: "Spacer", icon: "Flex" },
    ],
  },
  {
    id: "components",
    label: "Components",
    items: [
      { type: "section", label: "Section", icon: "Box" },
      { type: "tabs", label: "Tabs", icon: "Tabs" },
      { type: "table", label: "Table", icon: "Table" },
      { type: "list", label: "List", icon: "List" },
      { type: "tree", label: "Tree", icon: "Tree" },
      { type: "calendar", label: "Calendar", icon: "Calendar" },
      { type: "badge", label: "Badge", icon: "Badge" },
      { type: "image", label: "Image", icon: "Image" },
      { type: "custom", label: "Custom", icon: "Custom" },
    ],
  },
  {
    id: "buttons",
    label: "Buttons",
    items: [
      { type: "button", label: "Button", icon: "Button" },
      { type: "checkbox", label: "Check Box", icon: "Check" },
      { type: "radio", label: "Radio Button", icon: "Radio" },
      { type: "toggle", label: "Toggle", icon: "Toggle" },
    ],
  },
  {
    id: "inputs",
    label: "Input Widgets",
    items: [
      { type: "label", label: "Label", icon: "Text" },
      { type: "input", label: "Input", icon: "Input" },
      { type: "textarea", label: "Text Area", icon: "TextArea" },
      { type: "combobox", label: "Combo Box", icon: "Select" },
      { type: "slider", label: "Slider", icon: "Slider" },
    ],
  },
];

export const layoutTypes = new Set(["vstack", "hstack", "grid", "splitter", "spacer"]);
export const containerTypes = new Set([...rootSurfaceTypes, "vstack", "hstack", "grid", "splitter", "section", "tabs"]);
export const leafTypes = new Set(["button", "checkbox", "radio", "toggle", "label", "input", "textarea", "combobox", "slider", "table", "image", "custom", "list", "tree", "calendar", "badge", "spacer"]);

export function fieldsFor(node) {
  switch (node.type) {
    case "browser":
      return ["id", "title", "address"];
    case "window":
    case "mobile":
    case "dialog":
    case "section":
      return ["id", "title"];
    case "menu":
      return ["id"];
    case "grid":
      return ["id"];
    case "splitter":
      return ["id", "orientation"];
    case "tabs":
      return ["id"];
    case "button":
      return ["id", "label", "badge", "action", "anchor"];
    case "checkbox":
    case "radio":
    case "toggle":
      return ["id", "label", "action", "anchor"];
    case "input":
    case "textarea":
      return ["id", "label", "hint"];
    case "combobox":
      return ["id", "label", "hint"];
    case "slider":
      return ["id", "label"];
    case "label":
    case "image":
    case "custom":
      return ["id", "name", "purpose"];
    case "list":
    case "tree":
    case "calendar":
    case "badge":
      return ["id", "label"];
    default:
      return ["id"];
  }
}

export function canHaveChildren(node) {
  return Boolean(node && containerTypes.has(node.type));
}

export function isRootSurface(type) {
  return rootSurfaceTypes.includes(type);
}
