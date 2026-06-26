import { useEffect, useState } from "react";
import {
  clearDraft,
  deleteSavedDocument,
  findSavedDocumentByName,
  loadDraft,
  loadSavedDocument,
  loadSavedIndex,
  renameSavedDocument,
  saveDraft,
  saveNamedDocument,
} from "../model/browser-storage.js";
import { createHostFileApi } from "../model/host-file-api.js";
import { createShareUrl, decodeSharePayload } from "../model/share-url.js";
import { serializeSourceDocument } from "../model/source-document.js";
import { useEditorStore } from "../state/useEditorStore.js";
import { Canvas } from "./Canvas.jsx";
import { Inspector } from "./Inspector.jsx";
import { SourceEditor } from "./SourceEditor.jsx";
import { ToolPalette } from "./ToolPalette.jsx";
import { TreePanel } from "./TreePanel.jsx";

export function App() {
  const resetRoot = useEditorStore((state) => state.resetRoot);
  const replaceDocumentFromSource = useEditorStore((state) => state.replaceDocumentFromSource);
  const markSaved = useEditorStore((state) => state.markSaved);
  const undo = useEditorStore((state) => state.undo);
  const redo = useEditorStore((state) => state.redo);
  const canUndo = useEditorStore((state) => state.undoStack.length > 0);
  const canRedo = useEditorStore((state) => state.redoStack.length > 0);
  const mode = useEditorStore((state) => state.mode);
  const dirty = useEditorStore((state) => state.dirty);
  const documentName = useEditorStore((state) => state.documentName);
  const savedDocumentId = useEditorStore((state) => state.savedDocumentId);
  const currentFile = useEditorStore((state) => state.currentFile);
  const revision = useEditorStore((state) => state.revision);
  const root = useEditorStore((state) => state.root);
  const source = useEditorStore((state) => state.currentSource());
  const [hostFileApi] = useState(() => createHostFileApi());
  const [savedDocs, setSavedDocs] = useState([]);
  const [projectFiles, setProjectFiles] = useState([]);
  const [message, setMessage] = useState("");
  const [editorMode, setEditorMode] = useState("visual");

  useEffect(() => {
    window.uisketchHasUnsavedChanges = () => useEditorStore.getState().dirty;
    window.uisketchConfirmClose = () => confirmDiscard();
    return () => {
      delete window.uisketchHasUnsavedChanges;
      delete window.uisketchConfirmClose;
    };
  }, []);

  useEffect(() => {
    hostFileApi.setDirty?.(dirty);
  }, [dirty, hostFileApi]);

  useEffect(() => {
    if (!dirty) return;
    const handler = (event) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", handler);
    return () => window.removeEventListener("beforeunload", handler);
  }, [dirty]);

  useEffect(() => {
    if (mode !== "browser") return;
    setSavedDocs(loadSavedIndex());
    const sharedSource = decodeShareHash();
    if (sharedSource) {
      replaceDocumentFromSource(sharedSource, { documentName: "Shared Sketch" });
      return;
    }
    const draft = loadDraft();
    if (draft?.source && confirm("Restore the autosaved draft?")) {
      replaceDocumentFromSource(draft.source, { documentName: draft.documentName || "Restored Draft" });
    }
  }, [mode, replaceDocumentFromSource]);

  useEffect(() => {
    if (mode !== "browser" || !dirty) return;
    saveDraft({ source, documentName });
  }, [dirty, documentName, mode, source]);

  useEffect(() => {
    if (mode !== "local") return;
    refreshProjectFiles();
    const initialFile = hostFileApi.config().initialFile;
    if (initialFile) openProjectFile(initialFile, { skipConfirm: true });
  }, [mode, hostFileApi]);

  function confirmDiscard() {
    return !useEditorStore.getState().dirty || confirm("Unsaved changes will be discarded. Continue?");
  }

  function newRoot(type) {
    if (!confirmDiscard()) return;
    resetRoot(type);
  }

  function saveBrowserDocument() {
    const current = useEditorStore.getState();
    const defaultName = current.documentName || current.root.title || "Untitled Sketch";
    const name = prompt("Save as", defaultName);
    if (!name) return;
    const existing = findSavedDocumentByName(name);
    if (existing && existing.id !== current.savedDocumentId && !confirm(`Overwrite "${name}"?`)) return;
    const record = saveNamedDocument({
      id: existing?.id || current.savedDocumentId || undefined,
      name,
      source: current.currentSource(),
    });
    clearDraft();
    markSaved({ documentName: record.name, savedDocumentId: record.id });
    setSavedDocs(loadSavedIndex());
    setMessage(`Saved "${record.name}"`);
  }

  function loadBrowserDocument(id) {
    if (!id || !confirmDiscard()) return;
    const record = loadSavedDocument(id);
    if (!record) return;
    replaceDocumentFromSource(record.source, { documentName: record.name, savedDocumentId: record.id });
    setMessage(`Loaded "${record.name}"`);
  }

  function deleteBrowserDocument(id) {
    if (!id) return;
    const record = loadSavedDocument(id);
    if (!record || !confirm(`Delete "${record.name}"?`)) return;
    deleteSavedDocument(id);
    setSavedDocs(loadSavedIndex());
    if (id === savedDocumentId) markSaved({ savedDocumentId: null });
  }

  function renameBrowserDocument() {
    if (!savedDocumentId) return;
    const name = prompt("Rename saved document", documentName);
    if (!name) return;
    const existing = findSavedDocumentByName(name);
    if (existing && existing.id !== savedDocumentId) {
      setMessage(`"${name}" already exists`);
      return;
    }
    const record = renameSavedDocument(savedDocumentId, name);
    markSaved({ documentName: record.name, savedDocumentId: record.id });
    setSavedDocs(loadSavedIndex());
    setMessage(`Renamed "${record.name}"`);
  }

  function copyShareUrl() {
    navigator.clipboard?.writeText(createShareUrl(window.location.href, source));
    setMessage("Share URL copied");
  }

  function decodeShareHash() {
    const hash = window.location.hash.replace(/^#/, "");
    const params = new URLSearchParams(hash);
    const value = params.get("s");
    if (!value) return "";
    try {
      return decodeSharePayload(value);
    } catch {
      setMessage("Could not decode share URL");
      return "";
    }
  }

  async function refreshProjectFiles() {
    try {
      const payload = await hostFileApi.listFiles();
      setProjectFiles(payload.files || []);
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function openProjectFile(path, options = {}) {
    if (!path || (!options.skipConfirm && !confirmDiscard())) return;
    try {
      const payload = await hostFileApi.readFile(path);
      replaceDocumentFromSource(payload.source, {
        documentName: payload.path,
        currentFile: payload.path,
        revision: payload.revision,
      });
      setMessage(`Opened ${payload.path}`);
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function saveProjectFile() {
    try {
      const current = useEditorStore.getState();
      if (!current.currentFile) {
        await newProjectFile({ skipConfirm: true });
        return;
      }
      const payload = await hostFileApi.writeFile({
        path: current.currentFile,
        source: current.currentSource(),
        revision: current.revision,
      });
      markSaved({ currentFile: payload.path, revision: payload.revision, documentName: payload.path });
      await refreshProjectFiles();
      setMessage(`Saved ${payload.path}`);
    } catch (error) {
      setMessage(error.message);
    }
  }

  async function newProjectFile(options = {}) {
    if (!options.skipConfirm && !confirmDiscard()) return;
    const path = prompt("New project file", currentFile || "sketch.uisketch.md");
    if (!path) return;
    try {
      const source = serializeSourceDocument(root, path);
      const payload = await hostFileApi.createFile({ path, source });
      replaceDocumentFromSource(payload.source, {
        documentName: payload.path,
        currentFile: payload.path,
        revision: payload.revision,
        savePointSource: payload.source,
      });
      await refreshProjectFiles();
      setMessage(`Created ${payload.path}`);
    } catch (error) {
      setMessage(error.message);
    }
  }

  return (
    <>
      <header className="navbar min-h-16 border-b border-base-300 bg-base-100 px-4">
        <div className="flex-1">
          <div>
            <h1 className="text-xl font-bold">UI Sketch Editor</h1>
            <p className="text-sm text-base-content/60">
              {mode === "local" ? `Local project${currentFile ? `: ${currentFile}` : ""}` : documentName}
              {dirty ? " *" : ""}
              {message ? ` - ${message}` : ""}
            </p>
          </div>
        </div>
        <div className="flex flex-wrap justify-end gap-2">
          {mode === "browser" ? (
            <>
              <button className="btn btn-sm btn-primary" onClick={saveBrowserDocument}>
                Save
              </button>
              <select className="select select-sm w-44" onChange={(event) => loadBrowserDocument(event.target.value)} value="">
                <option value="">Load saved...</option>
                {savedDocs.map((doc) => (
                  <option key={doc.id} value={doc.id}>
                    {doc.name}
                  </option>
                ))}
              </select>
              <button className="btn btn-sm btn-ghost" disabled={!savedDocumentId} onClick={() => deleteBrowserDocument(savedDocumentId)}>
                Delete
              </button>
              <button className="btn btn-sm btn-ghost" disabled={!savedDocumentId} onClick={renameBrowserDocument}>
                Rename
              </button>
              <button className="btn btn-sm btn-ghost" onClick={copyShareUrl}>
                Share
              </button>
            </>
          ) : (
            <>
              <button className="btn btn-sm btn-primary" onClick={saveProjectFile}>
                Save
              </button>
              <button className="btn btn-sm btn-ghost" onClick={newProjectFile}>
                New file
              </button>
              <select className="select select-sm w-52" onChange={(event) => openProjectFile(event.target.value)} value="">
                <option value="">Open project file...</option>
                {projectFiles.map((file) => (
                  <option key={file.path} value={file.path}>
                    {file.path}
                  </option>
                ))}
              </select>
            </>
          )}
          <button className="btn btn-sm" disabled={!canUndo} onClick={undo}>
            Undo
          </button>
          <button className="btn btn-sm" disabled={!canRedo} onClick={redo}>
            Redo
          </button>
          <button className="btn btn-sm btn-ghost" onClick={() => newRoot("browser")}>
            New browser
          </button>
          <button className="btn btn-sm btn-ghost" onClick={() => newRoot("window")}>
            New window
          </button>
        </div>
      </header>
      <main className="min-w-[1160px] bg-base-200 p-3">
        <div role="tablist" className="tabs tabs-lift">
          <button className={`tab ${editorMode === "visual" ? "tab-active" : ""}`} role="tab" type="button" onClick={() => setEditorMode("visual")}>
            Visual Editor
          </button>
          <button className={`tab ${editorMode === "source" ? "tab-active" : ""}`} role="tab" type="button" onClick={() => setEditorMode("source")}>
            Source&Preview
          </button>
        </div>
        {editorMode === "visual" ? (
          <section className="grid grid-cols-[340px_minmax(520px,1fr)_320px] gap-3" data-testid="visual-mode">
            <aside className="card grid h-[calc(100vh-124px)] min-h-0 grid-rows-[minmax(220px,1fr)_minmax(180px,0.7fr)] overflow-hidden bg-base-100">
              <div className="min-h-0 overflow-auto">
                <ToolPalette />
              </div>
              <TreePanel />
            </aside>
            <section className="card h-[calc(100vh-124px)] min-h-0 bg-base-100">
              <div className="pane-title">Visual Editor</div>
              <Canvas />
            </section>
            <aside className="card grid h-[calc(100vh-124px)] min-h-0 grid-rows-[auto_1fr] overflow-hidden bg-base-100">
              <div className="pane-title">Properties</div>
              <div className="overflow-auto">
                <Inspector />
              </div>
            </aside>
          </section>
        ) : (
          <SourceEditor hostFileApi={hostFileApi} onMessage={setMessage} />
        )}
      </main>
    </>
  );
}
