import assert from "node:assert/strict";
import { deleteSavedDocument, loadSavedIndex, saveNamedDocument } from "../src/model/browser-storage.js";
import { paletteGroups } from "../src/model/component-catalog.js";
import { createShareUrl, decodeSharePayload } from "../src/model/share-url.js";
import { parseSourceDocument, serializeSourceDocument } from "../src/model/source-document.js";
import { parseLayoutYaml, serializeYaml } from "../src/model/yaml.js";
import { useEditorStore } from "../src/state/useEditorStore.js";

const store = useEditorStore.getState();

assert.equal(paletteGroups.some((group) => group.id === "surfaces"), false);
const paletteTypes = paletteGroups.flatMap((group) => group.items.map((item) => item.type));
assert.equal(paletteTypes.includes("toggle"), true);
assert.equal(paletteTypes.includes("textarea"), true);
assert.equal(paletteTypes.includes("combobox"), true);
assert.equal(paletteTypes.includes("splitter"), true);

store.insertNewNode("button", store.selectedPath);
let state = useEditorStore.getState();
assert.deepEqual(state.selectedPath, [2]);

state.updateSelectedField("label", "Submit order");
state.updateSelectedField("note", "Visible annotation");
state.updateSelectedField("prompt", "AI note for the button");
state.updateSelectedField("data", { role: "primary", tracking: { event: "submit_order" }, flags: ["default"] });
state = useEditorStore.getState();

let yaml = serializeYaml(state.root);
assert.match(yaml, /browser:/);
assert.match(yaml, /button:/);
assert.match(yaml, /label: Submit order/);
assert.match(yaml, /note: Visible annotation/);
assert.match(yaml, /prompt: AI note for the button/);
assert.match(yaml, /data:\n\s+role: primary\n\s+tracking:\n\s+event: submit_order\n\s+flags:\n\s+- default/);
assert.match(yaml, /children:\n\s+- hstack:/);
assert.deepEqual(parseLayoutYaml(yaml).children[2].data, { flags: ["default"], role: "primary", tracking: { event: "submit_order" } });
assert.equal(state.undoStack.length > 0, true);

state.undo();
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.equal(yaml.includes("submit_order"), false);
assert.equal(state.redoStack.length > 0, true);

state.redo();
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /label: Submit order/);
assert.match(yaml, /prompt: AI note for the button/);
assert.match(yaml, /data:\n\s+role: primary\n\s+tracking:\n\s+event: submit_order\n\s+flags:\n\s+- default/);

state.replaceRootType("mobile");
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /mobile:/);
assert.match(yaml, /menu:\n\s+- Home\n\s+- Search\n\s+- Settings/);
assert.equal(yaml.includes("address:"), false);

state.replaceRootType("menu");
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /menu:/);
const menuRootHeader = yaml.split("\n  children:")[0];
assert.equal(menuRootHeader.includes("title:"), false);
assert.equal(menuRootHeader.includes("address:"), false);

state.replaceRootType("browser");
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /address: "https:\/\/shibukawa.github.io\/uisketch"/);

state.resetRoot("window");
state = useEditorStore.getState();
state.updateSelectedField("menu", ["File", "Edit"]);
state.insertNewNode("button", []);
yaml = serializeYaml(useEditorStore.getState().root);
assert.match(yaml, /window:/);
assert.match(yaml, /menu:\n\s+- File\n\s+- Edit/);
assert.match(yaml, /children:\n\s+- button:/);

state.replaceRootType("dialog");
state = useEditorStore.getState();
state.updateSelectedField("buttons", [
  { type: "button", label: "Cancel" },
  { type: "button", label: "OK", action: "action.confirm" },
]);
yaml = serializeYaml(useEditorStore.getState().root);
assert.match(yaml, /dialog:/);
assert.match(yaml, /buttons:\n\s+- button:\n\s+label: Cancel\n\s+- button:\n\s+label: OK\n\s+action: action.confirm/);
assert.equal(parseLayoutYaml(yaml).buttons[1].action, "action.confirm");

state.setSelectedPath([]);
state.insertNewNode("grid", state.selectedPath);
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /grid:/);
assert.match(yaml, /columns: 2/);

state.insertNewNode("tabs", state.selectedPath);
state = useEditorStore.getState();
yaml = serializeYaml(state.root);
assert.match(yaml, /tabs:/);
assert.match(yaml, /labels:\n\s+- \[Tab A\]/);

state.setSelectedPath([]);
state.insertNewNode("toggle", []);
state.insertNewNode("textarea", []);
state.insertNewNode("combobox", []);
yaml = serializeYaml(useEditorStore.getState().root);
assert.match(yaml, /toggle:/);
assert.match(yaml, /textarea:/);
assert.match(yaml, /combobox:/);

state.setSelectedPath([]);
const beforeSpacer = serializeYaml(useEditorStore.getState().root);
state.insertNewNode("spacer", []);
assert.equal(serializeYaml(useEditorStore.getState().root), beforeSpacer);

state.setSelectedPath([]);
state.insertNewNode("hstack", []);
state = useEditorStore.getState();
state.insertNewNode("spacer", state.selectedPath);
yaml = serializeYaml(useEditorStore.getState().root);
assert.match(yaml, /spacer:/);

state.setSelectedPath([]);
state.insertNewNode("vstack", []);
state = useEditorStore.getState();
state.insertNewNode("spacer", state.selectedPath);
yaml = serializeYaml(useEditorStore.getState().root);
assert.match(yaml, /vstack:\n\s+children:\n\s+- spacer:/);

state = useEditorStore.getState();
state.setDragging({ kind: "new", type: "button" });
state = useEditorStore.getState();
state.setHoveredDropTarget({ parentPath: [], index: 1 });
assert.deepEqual(useEditorStore.getState().hoveredDropTarget, { parentPath: [], index: 1 });
useEditorStore.getState().setDragging(null);
assert.equal(useEditorStore.getState().hoveredDropTarget, null);

const roundTripRoot = parseLayoutYaml(yaml);
assert.equal(roundTripRoot.type, useEditorStore.getState().root.type);
assert.equal(roundTripRoot.children.some((child) => child.type === "hstack"), true);

const source = serializeSourceDocument(roundTripRoot, "Smoke Sketch");
const parsedSourceRoot = parseSourceDocument(source);
assert.equal(parsedSourceRoot.type, roundTripRoot.type);
assert.equal(parsedSourceRoot.id, roundTripRoot.id);

const shareUrl = createShareUrl("https://example.test/editor?mode=local&file=secret.uisketch.md#old=1", source);
assert.equal(shareUrl.includes("mode=local"), false);
assert.equal(shareUrl.includes("secret.uisketch.md"), false);
const sharePayload = new URL(shareUrl).hash.slice(3);
assert.equal(decodeSharePayload(sharePayload), source);

const memoryStorage = new Map();
const storage = {
  getItem(key) {
    return memoryStorage.has(key) ? memoryStorage.get(key) : null;
  },
  setItem(key, value) {
    memoryStorage.set(key, value);
  },
  removeItem(key) {
    memoryStorage.delete(key);
  },
};
saveNamedDocument({ id: "doc-1", name: "Smoke", source }, storage);
assert.equal(loadSavedIndex(storage)[0].name, "Smoke");
deleteSavedDocument("doc-1", storage);
assert.equal(loadSavedIndex(storage).length, 0);

state = useEditorStore.getState();
state.replaceDocumentFromSource(source, { documentName: "Smoke Sketch", savedDocumentId: "doc-1" });
assert.equal(useEditorStore.getState().dirty, false);
useEditorStore.getState().insertNewNode("button", []);
assert.equal(useEditorStore.getState().dirty, true);
