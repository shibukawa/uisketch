import { Fragment } from "react";
import { canHaveChildren, containerTypes, layoutTypes, leafTypes } from "../model/component-catalog.js";
import { leafText } from "../model/document.js";
import { samePath } from "../model/tree.js";
import { useEditorStore } from "../state/useEditorStore.js";

export function Canvas() {
  const root = useEditorStore((state) => state.root);
  return (
    <div className="canvas-grid min-h-0 flex-1 overflow-auto p-4">
      <CanvasNode node={root} path={[]} />
    </div>
  );
}

function CanvasNode({ node, path, parentType = "" }) {
  const selectedPath = useEditorStore((state) => state.selectedPath);
  const setSelectedPath = useEditorStore((state) => state.setSelectedPath);
  const setDragging = useEditorStore((state) => state.setDragging);
  const deleteAtPath = useEditorStore((state) => state.deleteAtPath);
  const selected = samePath(path, selectedPath);
  const layout = layoutTypes.has(node.type);
  const root = path.length === 0;
  const dialogButton = path[0] === "buttons";

  return (
    <article
      draggable={!root && !dialogButton}
      className={`node-card node-${node.type} ${containerTypes.has(node.type) ? "container" : "leaf"} ${layout ? "node-layout" : "node-component"} ${selected ? "node-selected" : ""}`}
      onClick={(event) => {
        event.stopPropagation();
        setSelectedPath(path);
      }}
      onDragStart={(event) => {
        event.stopPropagation();
        if (!path.length) return;
        setDragging({ kind: "move", path });
        event.dataTransfer.effectAllowed = "move";
        event.dataTransfer.setData("text/plain", path.join("."));
      }}
      onDragEnd={() => setDragging(null)}
    >
      {!root && (
        <button
          type="button"
          className="node-delete btn btn-error btn-xs"
          aria-label={`Delete ${node.type}`}
          onClick={(event) => {
            event.stopPropagation();
            deleteAtPath(path);
          }}
        >
          ×
        </button>
      )}
      {node.type === "browser" && <BrowserChrome address={node.address || "https://example.test"} />}
      {node.type === "window" && <WindowChrome title={node.title || "Window"} />}
      {node.type === "window" && node.menu?.length > 0 && <SurfaceMenu items={node.menu} position="top" />}
      {node.type === "mobile" && <ChromeLine>Mobile viewport</ChromeLine>}
      {node.type === "dialog" && <DialogChrome title={node.title || "Dialog"} />}
      {node.type === "section" && node.title && <SectionLegend>{node.title}</SectionLegend>}
      {node.type === "tabs" && <TabsStrip node={node} />}
      {node.badge && <span className="node-badge" title={`Badge ${node.badge}`}>{badgeText(node.badge)}</span>}
      {node.note && <span className="node-note" title={node.note}>note</span>}
      {leafTypes.has(node.type) && <LeafBody node={node} parentType={parentType} />}
      {canHaveChildren(node) && <Children node={node} path={path} />}
      {node.type === "dialog" && <DialogButtons node={node} />}
      {node.type === "mobile" && node.menu?.length > 0 && <SurfaceMenu items={node.menu} position="bottom" />}
    </article>
  );
}

function ChromeLine({ children }) {
  return <div className="m-3 rounded-full border border-base-300 bg-base-200 px-3 py-2 text-sm text-base-content/60">{children}</div>;
}

function BrowserChrome({ address }) {
  return (
    <div className="surface-chrome browser-chrome">
      <div className="chrome-controls chrome-controls-left">
        <span>‹</span>
        <span>›</span>
        <span>↻</span>
      </div>
      <div className="chrome-address">{address}</div>
      <WindowButtons count={3} />
    </div>
  );
}

function WindowChrome({ title }) {
  return (
    <div className="surface-chrome window-chrome">
      <span className="chrome-title">{title}</span>
      <WindowButtons count={3} />
    </div>
  );
}

function DialogChrome({ title }) {
  return (
    <div className="surface-chrome dialog-chrome">
      <span className="chrome-title">{title}</span>
      <WindowButtons count={1} />
    </div>
  );
}

function SurfaceMenu({ items, position }) {
  return (
    <nav className={`surface-menu surface-menu-${position}`}>
      {items.map((item, index) => (
        <span key={`${item}-${index}`}>{item}</span>
      ))}
    </nav>
  );
}

function DialogButtons({ node }) {
  const buttons = node.buttons || [];
  if (!buttons.length) return null;
  return (
    <div className="dialog-buttons-row">
      {buttons.map((button, index) => (
        <CanvasNode key={index} node={button} path={["buttons", index]} parentType="dialog-buttons" />
      ))}
    </div>
  );
}

function WindowButtons({ count }) {
  return (
    <div className="chrome-controls">
      {Array.from({ length: count }).map((_, index) => (
        <span key={index} />
      ))}
    </div>
  );
}

function badgeText(value) {
  return String(value).trim();
}

function SectionLegend({ children }) {
  return <div className="section-legend">{children}</div>;
}

function TabsStrip({ node }) {
  return (
    <div className="tabs-strip m-3 flex gap-1 border-b border-base-300">
      {(node.labels || []).map((label, index) => (
        <span key={index} className={`tab tab-sm ${label.selected ? "tab-active" : ""}`}>
          {label.text}
        </span>
      ))}
    </div>
  );
}

function LeafBody({ node, parentType }) {
  if (node.type === "button") return <button className="sketch-control sketch-button">{leafText(node)}</button>;
  if (node.type === "checkbox") return <label className="sketch-choice"><span className="choice-box choice-check" />{leafText(node)}</label>;
  if (node.type === "radio") return <label className="sketch-choice"><span className="choice-box choice-radio" />{leafText(node)}</label>;
  if (node.type === "toggle") return <label className="sketch-choice"><span className="choice-toggle"><span /></span>{leafText(node)}</label>;
  if (node.type === "input") {
    return (
      <label className="sketch-field">
        {node.label && <span className="text-sm font-medium">{node.label}</span>}
        <input className="sketch-input" placeholder={node.hint || ""} readOnly />
      </label>
    );
  }
  if (node.type === "textarea") {
    return (
      <label className="sketch-field">
        {node.label && <span className="text-sm font-medium">{node.label}</span>}
        <textarea className="sketch-input sketch-textarea" placeholder={node.hint || ""} readOnly />
      </label>
    );
  }
  if (node.type === "combobox") {
    return (
      <label className="sketch-field">
        {node.label && <span className="text-sm font-medium">{node.label}</span>}
        <select className="sketch-input" value={node.options?.[0] || ""} onChange={() => {}}>
          {(node.options?.length ? node.options : ["Option"]).map((option) => (
            <option key={option}>{option}</option>
          ))}
        </select>
      </label>
    );
  }
  if (node.type === "slider") return <label className="sketch-field"><span className="text-sm font-medium">{leafText(node)}</span><input type="range" className="range range-sm" readOnly /></label>;
  if (node.type === "table") return <TableBody node={node} />;
  if (node.type === "image") return <div className="image-placeholder m-3 grid min-h-20 place-items-center rounded-btn border border-dashed border-base-300 text-base-content/45">{leafText(node)}</div>;
  if (node.type === "custom") return <div className="image-placeholder m-3 grid min-h-20 place-items-center rounded-btn border border-dashed border-base-300 text-base-content/45">{leafText(node)}</div>;
  if (node.type === "list") return <ul className="m-3 list-disc pl-6 text-sm"><li>{leafText(node)}</li></ul>;
  if (node.type === "tree") return <ul className="m-3 list-disc pl-6 text-sm"><li>{leafText(node)}</li><li className="ml-4 text-base-content/50">Child</li></ul>;
  if (node.type === "calendar") return <div className="m-3 grid grid-cols-7 gap-1 text-center text-xs text-base-content/60">{Array.from({ length: 14 }).map((_, index) => <span key={index} className="rounded border border-base-300 py-1">{index + 1}</span>)}</div>;
  if (node.type === "badge") return <span className="badge badge-outline m-3">{leafText(node)}</span>;
  if (node.type === "spacer") return <SpacerBody direction={parentType === "vstack" ? "vertical" : "horizontal"} />;
  return <div className="p-3 text-base-content">{leafText(node)}</div>;
}

function SpacerBody({ direction }) {
  return (
    <div className={`spacer-body spacer-${direction}`}>
      <span className="spacer-handle" />
      <span className="spacer-line" />
      <span className="spacer-handle" />
    </div>
  );
}

function TableBody({ node }) {
  const columns = node.columns?.length ? node.columns : ["Column"];
  return (
    <table className="sketch-table">
      <thead>
        <tr>
          {columns.map((column) => (
            <th key={column} className="border border-base-300 bg-base-200 px-3 py-2 text-left font-semibold">
              {column}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        <tr>
          {columns.map((column) => (
            <td key={column} className="border border-base-300 px-3 py-4 text-base-content/30">
              &nbsp;
            </td>
          ))}
        </tr>
      </tbody>
    </table>
  );
}

function Children({ node, path }) {
  const style = childStyle(node);
  const empty = !node.children?.length;
  const hasSpacer = Boolean(node.children?.some((child) => child.type === "spacer"));
  if (node.type === "splitter") {
    return (
      <div className={`children children-${node.type} ${empty ? "children-empty" : ""}`} style={style}>
        {(node.children || []).map((child, index) => (
          <CanvasNode key={index} node={child} path={[...path, index]} parentType={node.type} />
        ))}
      </div>
    );
  }
  return (
    <div className={`children children-${node.type} ${empty ? "children-empty" : ""} ${hasSpacer ? "children-has-spacer" : ""}`} style={style}>
      <DropZone parentPath={path} index={0} tail={empty} />
      {(node.children || []).map((child, index) => (
        <Fragment key={index}>
          <CanvasNode node={child} path={[...path, index]} parentType={node.type} />
          <DropZone parentPath={path} index={index + 1} tail={index === node.children.length - 1} fixedTail={hasSpacer && index === node.children.length - 1} />
        </Fragment>
      ))}
    </div>
  );
}

function childStyle(node) {
  if (node.type === "grid") return { gridTemplateColumns: `repeat(${Number(node.columns) || 2}, minmax(0, 1fr))` };
  if (node.type === "splitter") {
    const sizes = Array.isArray(node.sizes) && node.sizes.length === 2 ? node.sizes : [25, 75];
    const tracks = sizes.map((size) => (size === "$" || size === "*" ? "1fr" : `${Number(size) || 50}fr`)).join(" ");
    return node.orientation === "vertical" ? { gridTemplateRows: tracks } : { gridTemplateColumns: tracks };
  }
  return undefined;
}

function DropZone({ parentPath, index, tail, fixedTail = false }) {
  const dragging = useEditorStore((state) => state.dragging);
  const hoveredDropTarget = useEditorStore((state) => state.hoveredDropTarget);
  const setHoveredDropTarget = useEditorStore((state) => state.setHoveredDropTarget);
  const dropAt = useEditorStore((state) => state.dropAt);
  const hovered = Boolean(dragging && hoveredDropTarget && hoveredDropTarget.index === index && samePath(hoveredDropTarget.parentPath, parentPath));
  return (
    <div
      className={`drop-zone ${tail ? "drop-zone-tail" : ""} ${fixedTail ? "drop-zone-fixed-tail" : ""} ${hovered ? "drop-zone-active" : ""}`}
      onDragEnter={(event) => {
        event.preventDefault();
        event.stopPropagation();
        if (dragging) setHoveredDropTarget({ parentPath, index });
      }}
      onDragOver={(event) => {
        event.preventDefault();
        event.stopPropagation();
        if (dragging && !hovered) setHoveredDropTarget({ parentPath, index });
      }}
      onDragLeave={(event) => {
        event.stopPropagation();
        if (hovered) setHoveredDropTarget(null);
      }}
      onDrop={(event) => {
        event.preventDefault();
        event.stopPropagation();
        dropAt(parentPath, index);
      }}
    >
      <span>Insert here</span>
    </div>
  );
}
