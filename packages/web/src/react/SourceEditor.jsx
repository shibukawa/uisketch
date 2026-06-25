import { useEffect, useMemo, useRef, useState } from "react";
import { completionsForContext, componentTypes, propertyTypes, sourceSchemaFindings } from "../model/source-schema.js";
import { renderWithWasm } from "../model/wasm-render-api.js";
import { parseLayoutYaml, serializeYaml } from "../model/yaml.js";
import { useEditorStore } from "../state/useEditorStore.js";
import { Canvas } from "./Canvas.jsx";

export function SourceEditor({ hostFileApi, onMessage }) {
  const root = useEditorStore((state) => state.root);
  const mode = useEditorStore((state) => state.mode);
  const applyLayoutYamlEdit = useEditorStore((state) => state.applyLayoutYamlEdit);
  const canonicalYaml = useMemo(() => serializeYaml(root), [root]);
  const [draft, setDraft] = useState(canonicalYaml);
  const [staleDraft, setStaleDraft] = useState(false);
  const [previewTab, setPreviewTab] = useState("svg");
  const [cursor, setCursor] = useState({ line: 1, column: 1 });
  const [rendered, setRendered] = useState({ svg: "", text: "", findings: [], source: "local", error: "" });
  const editorRef = useRef(null);
  const lineRef = useRef(null);
  const highlightRef = useRef(null);

  useEffect(() => {
    if (staleDraft) return;
    setDraft(canonicalYaml);
  }, [canonicalYaml, staleDraft]);

  const parsedDraft = useMemo(() => {
    try {
      return { root: parseLayoutYaml(draft), error: "" };
    } catch (error) {
      return { root: null, error: error.message };
    }
  }, [draft]);

  const activeRoot = parsedDraft.root || root;
  const diagnostics = sourceDiagnostics(draft, parsedDraft.error, rendered.findings);
  const token = currentToken(draft, editorRef.current?.selectionStart || 0);
  const completions = completionsForContext(draft, editorRef.current?.selectionStart || 0).filter((keyword) => !token || keyword.startsWith(token)).slice(0, 24);

  useEffect(() => {
    let cancelled = false;
    renderPreview(draft, activeRoot, hostFileApi, mode).then((result) => {
      if (!cancelled) setRendered(result);
    });
    return () => {
      cancelled = true;
    };
  }, [draft, activeRoot, hostFileApi, mode]);

  function applyDraft() {
    try {
      applyLayoutYamlEdit(draft);
      setStaleDraft(false);
      onMessage?.("YAML source applied");
    } catch (error) {
      setStaleDraft(true);
      onMessage?.(`YAML source has errors: ${error.message}`);
    }
  }

  function formatDraft() {
    const next = serializeYaml(useEditorStore.getState().root);
    setDraft(next);
    setStaleDraft(false);
    updateCursor(next, editorRef.current?.selectionStart || 0);
  }

  function onDraftChange(value) {
    setDraft(value);
    setStaleDraft(true);
  }

  function syncScroll() {
    if (!editorRef.current) return;
    if (lineRef.current) lineRef.current.scrollTop = editorRef.current.scrollTop;
    if (highlightRef.current) {
      highlightRef.current.scrollTop = editorRef.current.scrollTop;
      highlightRef.current.scrollLeft = editorRef.current.scrollLeft;
    }
  }

  function updateCursor(value = draft, offset = editorRef.current?.selectionStart || 0) {
    const before = value.slice(0, offset);
    const lines = before.split("\n");
    setCursor({ line: lines.length, column: lines.at(-1).length + 1 });
  }

  function insertCompletion(keyword) {
    const editor = editorRef.current;
    if (!editor) return;
    const current = currentToken(draft, editor.selectionStart);
    const start = editor.selectionStart - current.length;
    const suffix = shouldCompleteAsKey(draft, start) ? ": " : "";
    const next = draft.slice(0, start) + keyword + suffix + draft.slice(editor.selectionEnd);
    setDraft(next);
    setStaleDraft(true);
    requestAnimationFrame(() => {
      editor.focus();
      editor.selectionStart = editor.selectionEnd = start + keyword.length + suffix.length;
      updateCursor(next, editor.selectionStart);
    });
  }

  function onKeyDown(event) {
    if (event.key !== "Tab") return;
    event.preventDefault();
    const editor = event.currentTarget;
    const next = draft.slice(0, editor.selectionStart) + "  " + draft.slice(editor.selectionEnd);
    const cursorOffset = editor.selectionStart + 2;
    setDraft(next);
    setStaleDraft(true);
    requestAnimationFrame(() => {
      editor.selectionStart = editor.selectionEnd = cursorOffset;
      updateCursor(next, cursorOffset);
    });
  }

  async function copySvg() {
    await copyText(rendered.svg, "SVG copied", onMessage);
  }

  async function copyPng() {
    const dataUrl = await svgToPngDataUrl(rendered.svg);
    await copyPngDataUrl(dataUrl);
    onMessage?.("PNG copied");
  }

  function downloadSvg() {
    downloadText("uisketch-preview.svg", rendered.svg, "image/svg+xml");
    onMessage?.("SVG downloaded");
  }

  async function downloadPng() {
    const dataUrl = await svgToPngDataUrl(rendered.svg);
    downloadDataUrl("uisketch-preview.png", dataUrl);
    onMessage?.("PNG downloaded");
  }

  async function copyAscii() {
    await copyText(rendered.text, "ASCII copied", onMessage);
  }

  function downloadAscii() {
    downloadText("uisketch-preview.txt", rendered.text, "text/plain");
    onMessage?.("ASCII downloaded");
  }

  return (
    <div className="source-mode-grid">
      <section className="card min-h-0 overflow-hidden bg-base-100">
        <div className="source-editor-toolbar">
          <div>
            <div className="font-semibold">YAML Source</div>
            <div className="text-xs text-base-content/60">
              Ln {cursor.line}, Col {cursor.column}
              {staleDraft ? " - draft" : ""}
            </div>
          </div>
          <div className="flex gap-2">
            <button className="btn btn-sm" type="button" onClick={formatDraft}>
              Format
            </button>
            <button className="btn btn-sm btn-primary" type="button" onClick={applyDraft} disabled={Boolean(parsedDraft.error)}>
              Apply YAML
            </button>
          </div>
        </div>
        <div className="source-code-editor">
          <pre ref={lineRef} className="source-line-numbers" aria-hidden="true">
            {lineNumbers(draft)}
          </pre>
          <pre ref={highlightRef} className="source-highlight" aria-hidden="true" dangerouslySetInnerHTML={{ __html: highlightSource(draft) }} />
          <textarea
            ref={editorRef}
            className="source-textarea"
            value={draft}
            aria-label="YAML source editor"
            spellCheck="false"
            onChange={(event) => onDraftChange(event.target.value)}
            onClick={(event) => updateCursor(event.currentTarget.value, event.currentTarget.selectionStart)}
            onKeyDown={onKeyDown}
            onKeyUp={(event) => updateCursor(event.currentTarget.value, event.currentTarget.selectionStart)}
            onScroll={syncScroll}
          />
        </div>
        <div className="source-assist-grid">
          <section className="min-h-0 overflow-auto border-r border-base-300 p-3">
            <div className="mb-2 text-xs font-bold uppercase text-base-content/60">Completions</div>
            <div className="flex flex-wrap gap-2">
              {completions.map((keyword) => (
                <button key={keyword} className="btn btn-xs font-mono" type="button" onClick={() => insertCompletion(keyword)}>
                  {keyword}
                </button>
              ))}
            </div>
          </section>
          <section className="min-h-0 overflow-auto p-3">
            <div className="mb-2 text-xs font-bold uppercase text-base-content/60">Diagnostics</div>
            <pre className="source-diagnostics">{diagnostics.length ? diagnostics.map((finding) => `${finding.severity}: ${finding.message}`).join("\n") : "No diagnostics."}</pre>
          </section>
        </div>
      </section>

      <section className="card min-h-0 overflow-hidden bg-base-100">
        <div role="tablist" className="tabs tabs-lift border-b border-base-300 px-2 pt-2">
          {[
            ["svg", "Preview SVG"],
            ["ascii", "Preview ASCII"],
          ].map(([id, label]) => (
            <button key={id} className={`tab ${previewTab === id ? "tab-active" : ""}`} role="tab" type="button" onClick={() => setPreviewTab(id)}>
              {label}
            </button>
          ))}
          <span className="ml-auto self-center pr-2 text-xs text-base-content/50">render: {rendered.source}</span>
        </div>
        <div className="source-preview-pane">
          {previewTab === "svg" ? (
            <>
              <div className="source-preview-actions">
                <button className="btn btn-xs" type="button" onClick={copySvg} disabled={!rendered.svg}>
                  copy(svg)
                </button>
                <button className="btn btn-xs" type="button" onClick={copyPng} disabled={!rendered.svg}>
                  copy(png)
                </button>
                <button className="btn btn-xs" type="button" onClick={downloadSvg} disabled={!rendered.svg}>
                  download(svg)
                </button>
                <button className="btn btn-xs" type="button" onClick={downloadPng} disabled={!rendered.svg}>
                  download(png)
                </button>
              </div>
              <SvgPreview html={rendered.svg} root={activeRoot} />
            </>
          ) : null}
          {previewTab === "ascii" ? (
            <>
              <div className="source-preview-actions">
                <button className="btn btn-xs" type="button" onClick={copyAscii} disabled={!rendered.text}>
                  Copy
                </button>
                <button className="btn btn-xs" type="button" onClick={downloadAscii} disabled={!rendered.text}>
                  Download
                </button>
              </div>
              <pre className="source-preview-pre">{rendered.text}</pre>
            </>
          ) : null}
        </div>
      </section>
    </div>
  );
}

function SvgPreview({ html, root }) {
  if (html) return <div className="source-svg-preview" dangerouslySetInnerHTML={{ __html: html }} />;
  return (
    <div className="source-canvas-preview">
      <CanvasPreviewRoot root={root} />
    </div>
  );
}

function CanvasPreviewRoot({ root }) {
  const original = useEditorStore((state) => state.root);
  if (root === original) return <Canvas />;
  return <pre className="source-preview-pre">{renderAscii(root)}</pre>;
}

async function renderPreview(yaml, root, hostFileApi, mode) {
  if (!root) return { svg: "", text: "", findings: [], source: "none", error: "" };
  if (mode === "local" && hostFileApi?.render) {
    try {
      const result = await hostFileApi.render(yaml);
      return normalizeRenderResult(result, "go");
    } catch (error) {
      return { ...fallbackRender(root), source: "local-fallback", error: error.message };
    }
  }
  if (mode !== "local") {
    try {
      return normalizeRenderResult(await renderWithWasm(yaml), "wasm");
    } catch (error) {
      return { ...fallbackRender(root), source: "wasm-fallback", error: error.message };
    }
  }
  return { ...fallbackRender(root), source: "local-fallback", error: "" };
}

function normalizeRenderResult(result, source) {
  return {
    svg: result.svg || result.SVG || "",
    text: result.text || result.Text || "",
    findings: result.findings || result.Findings || [],
    source,
    error: result.error || result.Error || "",
  };
}

function fallbackRender(root) {
  return { svg: "", text: renderAscii(root), findings: [] };
}

function lineNumbers(source) {
  return Array.from({ length: Math.max(1, source.split("\n").length) }, (_, index) => index + 1).join("\n");
}

function currentToken(source, offset) {
  const match = source.slice(0, offset).match(/[A-Za-z][A-Za-z0-9_-]*$/);
  return match ? match[0] : "";
}

function shouldCompleteAsKey(source, start) {
  const lineStart = source.lastIndexOf("\n", Math.max(0, start - 1)) + 1;
  return /^\s*-?\s*$/.test(source.slice(lineStart, start));
}

function sourceDiagnostics(source, parseError, renderFindings = []) {
  const findings = [];
  if (parseError) findings.push({ severity: "Error", message: parseError });
  findings.push(...sourceSchemaFindings(source));
  for (const finding of renderFindings) {
    findings.push({ severity: normalizeSeverity(finding.severity || finding.Severity), message: finding.message || finding.Message || "" });
  }
  return findings.filter((finding) => finding.message);
}

function normalizeSeverity(value) {
  return String(value || "Warning").toLowerCase() === "error" ? "Error" : "Warning";
}

function highlightSource(source) {
  return source
    .split("\n")
    .map((line) => {
      const escaped = escapeHtml(line);
      if (/^\s*#/.test(line)) return `<span class="source-token-comment">${escaped}</span>`;
      return escaped.replace(/^(\s*-?\s*)([A-Za-z][A-Za-z0-9_-]*)(:)/, (_match, prefix, key, colon) => {
        const cls = componentTypes.includes(key) ? "source-token-component" : propertyTypes.includes(key) ? "source-token-key" : "source-token-string";
        return `${prefix}<span class="${cls}">${key}</span>${colon}`;
      });
    })
    .join("\n");
}

function renderAscii(root) {
  const lines = [];
  walkAscii(root, 0, lines);
  return lines.join("\n") + "\n";
}

async function copyText(value, message, onMessage) {
  await navigator.clipboard?.writeText(value || "");
  onMessage?.(message);
}

async function copyPngDataUrl(dataUrl) {
  const blob = await (await fetch(dataUrl)).blob();
  if (navigator.clipboard?.write && globalThis.ClipboardItem) {
    await navigator.clipboard.write([new ClipboardItem({ "image/png": blob })]);
    return;
  }
  await navigator.clipboard?.writeText(dataUrl);
}

function downloadText(filename, text, type) {
  const blob = new Blob([text || ""], { type });
  const url = URL.createObjectURL(blob);
  downloadUrl(filename, url);
  URL.revokeObjectURL(url);
}

function downloadDataUrl(filename, dataUrl) {
  downloadUrl(filename, dataUrl);
}

function downloadUrl(filename, url) {
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
}

function svgToPngDataUrl(svgText) {
  return new Promise((resolve, reject) => {
    const image = new Image();
    const svgBlob = new Blob([svgText || ""], { type: "image/svg+xml" });
    const url = URL.createObjectURL(svgBlob);
    image.onload = () => {
      const canvas = document.createElement("canvas");
      canvas.width = image.naturalWidth || image.width || 960;
      canvas.height = image.naturalHeight || image.height || 640;
      const ctx = canvas.getContext("2d");
      ctx.drawImage(image, 0, 0);
      URL.revokeObjectURL(url);
      resolve(canvas.toDataURL("image/png"));
    };
    image.onerror = () => {
      URL.revokeObjectURL(url);
      reject(new Error("failed to convert SVG to PNG"));
    };
    image.src = url;
  });
}

function walkAscii(node, depth, lines) {
  const pad = "  ".repeat(depth);
  const label = node.title || node.label || node.id || "";
  lines.push(`${pad}+ ${node.type}${label ? `: ${label}` : ""}`);
  for (const child of node.children || []) walkAscii(child, depth + 1, lines);
}

function escapeHtml(value) {
  return String(value || "").replace(/[&<>"']/g, (ch) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[ch]);
}
