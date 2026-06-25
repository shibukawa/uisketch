import { create } from "zustand";
import { canHaveChildren, isRootSurface } from "../model/component-catalog.js";
import { createDefaultNode, createInitialRoot, createRoot } from "../model/document.js";
import { parseSourceDocument, serializeSourceDocument } from "../model/source-document.js";
import { parseLayoutYaml } from "../model/yaml.js";
import { getNode, isDescendantPath, samePath } from "../model/tree.js";

function snapshot({ root, selectedPath, nextId, rootSurfaceCache }) {
  return {
    root: structuredClone(root),
    selectedPath: [...selectedPath],
    nextId,
    rootSurfaceCache: structuredClone(rootSurfaceCache || {}),
  };
}

function withHistory(state) {
  return {
    undoStack: [...state.undoStack, snapshot(state)].slice(-80),
    redoStack: [],
    dirty: true,
  };
}

function currentSource(state) {
  return serializeSourceDocument(state.root, state.documentName);
}

export const useEditorStore = create((set, get) => ({
  root: createInitialRoot(),
  selectedPath: [],
  dragging: null,
  hoveredDropTarget: null,
  rootSurfaceCache: {},
  nextId: 1,
  undoStack: [],
  redoStack: [],
  mode:
    new URLSearchParams(globalThis.location?.search || "").get("mode") === "local" ||
    Boolean(globalThis.uisketchHostFileApi || globalThis.go?.main?.DesktopApp)
      ? "local"
      : "browser",
  documentName: "Untitled Sketch",
  savedDocumentId: null,
  currentFile: null,
  revision: null,
  savePointSource: null,
  dirty: false,

  selectedNode() {
    return getNode(get().root, get().selectedPath);
  },

  setSelectedPath(path) {
    set({ selectedPath: path });
  },

  setDragging(dragging) {
    set({ dragging, hoveredDropTarget: null });
  },

  setHoveredDropTarget(target) {
    set({ hoveredDropTarget: target });
  },

  resetRoot(type) {
    set((state) => ({
      ...withHistory(state),
      root: createRoot(type),
      selectedPath: [],
      documentName: `${type[0].toUpperCase()}${type.slice(1)} Sketch`,
      savedDocumentId: null,
      currentFile: null,
      revision: null,
    }));
  },

  replaceDocument(root, metadata = {}) {
    set((state) => {
      const next = {
        ...state,
        root: structuredClone(root),
        selectedPath: [],
        dragging: null,
        hoveredDropTarget: null,
        undoStack: [],
        redoStack: [],
        documentName: metadata.documentName || state.documentName,
        savedDocumentId: metadata.savedDocumentId ?? null,
        currentFile: metadata.currentFile ?? null,
        revision: metadata.revision ?? null,
      };
      const source = metadata.savePointSource || serializeSourceDocument(next.root, next.documentName);
      return {
        ...next,
        savePointSource: source,
        dirty: false,
      };
    });
  },

  replaceDocumentFromSource(source, metadata = {}) {
    const root = parseSourceDocument(source);
    get().replaceDocument(root, { ...metadata, savePointSource: source });
  },

  applyLayoutYamlEdit(yaml) {
    const root = parseLayoutYaml(yaml);
    set((state) => ({
      ...withHistory(state),
      root,
      selectedPath: [],
      dragging: null,
      hoveredDropTarget: null,
    }));
  },

  markSaved(metadata = {}) {
    set((state) => {
      const next = {
        ...state,
        documentName: metadata.documentName || state.documentName,
        savedDocumentId: Object.hasOwn(metadata, "savedDocumentId") ? metadata.savedDocumentId : state.savedDocumentId,
        currentFile: Object.hasOwn(metadata, "currentFile") ? metadata.currentFile : state.currentFile,
        revision: Object.hasOwn(metadata, "revision") ? metadata.revision : state.revision,
      };
      return {
        ...next,
        savePointSource: currentSource(next),
        dirty: false,
      };
    });
  },

  currentSource() {
    return currentSource(get());
  },

  replaceRootType(type) {
    set((state) => {
      if (!isRootSurface(type) || state.root.type === type) return state;
      const rootSurfaceCache = {
        ...state.rootSurfaceCache,
        [state.root.type]: {
          title: state.root.title,
          address: state.root.address,
        },
      };
      const cached = rootSurfaceCache[type] || {};
      const root = createRoot(type);
      root.id = state.root.id || root.id;
      root.children = structuredClone(state.root.children || []);
      if (type !== "menu") root.title = cached.title || state.root.title || root.title;
      else delete root.title;
      if (type === "browser") root.address = cached.address || state.root.address || root.address;
      return {
        ...withHistory(state),
        rootSurfaceCache,
        root,
        selectedPath: [],
      };
    });
  },

  insertNewNode(type, targetPath, mode = "inside", index = null) {
    set((state) => {
      if (isRootSurface(type)) return state;
      const root = structuredClone(state.root);
      let parentPath = targetPath;
      let parent = getNode(root, parentPath);
      if (!canHaveChildren(parent)) {
        parentPath = targetPath.slice(0, -1);
        parent = getNode(root, parentPath);
      }
      if (!parent) return state;
      if (type === "spacer" && parent.type !== "hstack" && parent.type !== "vstack") return state;
      const node = createDefaultNode(type, state.nextId);
      parent.children ||= [];
      const insertIndex = mode === "at" && index != null ? index : parent.children.length;
      parent.children.splice(insertIndex, 0, node);
      return {
        ...withHistory(state),
        root,
        nextId: state.nextId + 1,
        selectedPath: [...parentPath, insertIndex],
      };
    });
  },

  moveNode(fromPath, toParentPath, toIndex) {
    set((state) => {
      if (!fromPath.length) return state;
      const root = structuredClone(state.root);
      const fromParent = getNode(root, fromPath.slice(0, -1));
      const movingIndex = fromPath[fromPath.length - 1];
      const moving = fromParent?.children?.[movingIndex];
      const toParent = getNode(root, toParentPath);
      if (!moving || !canHaveChildren(toParent) || isDescendantPath(toParentPath, fromPath)) return state;
      fromParent.children.splice(movingIndex, 1);
      let insertIndex = toIndex;
      if (samePath(fromPath.slice(0, -1), toParentPath) && movingIndex < toIndex) insertIndex -= 1;
      toParent.children.splice(insertIndex, 0, moving);
      return {
        ...withHistory(state),
        root,
        selectedPath: [...toParentPath, insertIndex],
      };
    });
  },

  dropAt(parentPath, index) {
    const dragging = get().dragging;
    set({ dragging: null, hoveredDropTarget: null });
    if (!dragging) return;
    if (dragging.kind === "new") get().insertNewNode(dragging.type, parentPath, "at", index);
    if (dragging.kind === "move") get().moveNode(dragging.path, parentPath, index);
  },

  deleteSelected() {
    get().deleteAtPath(get().selectedPath);
  },

  deleteAtPath(path) {
    set((state) => {
      if (!path.length) return state;
      const root = structuredClone(state.root);
      const parent = getNode(root, path.slice(0, -1));
      const index = path[path.length - 1];
      if (!parent?.children) return state;
      parent.children.splice(index, 1);
      return {
        ...withHistory(state),
        root,
        selectedPath: path.slice(0, -1),
      };
    });
  },

  updateSelectedField(key, value) {
    set((state) => {
      const root = structuredClone(state.root);
      const node = getNode(root, state.selectedPath);
      if (!node) return state;
      node[key] = value;
      return {
        ...withHistory(state),
        root,
      };
    });
  },

  undo() {
    set((state) => {
      const previous = state.undoStack.at(-1);
      if (!previous) return state;
      return {
        ...previous,
        undoStack: state.undoStack.slice(0, -1),
        redoStack: [...state.redoStack, snapshot(state)],
        dragging: null,
        dirty: true,
      };
    });
  },

  redo() {
    set((state) => {
      const next = state.redoStack.at(-1);
      if (!next) return state;
      return {
        ...next,
        undoStack: [...state.undoStack, snapshot(state)],
        redoStack: state.redoStack.slice(0, -1),
        dragging: null,
        dirty: true,
      };
    });
  },
}));
