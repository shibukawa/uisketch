import { displayName } from "../model/document.js";
import { samePath, walk } from "../model/tree.js";
import { useEditorStore } from "../state/useEditorStore.js";

export function TreePanel() {
  const root = useEditorStore((state) => state.root);
  const selectedPath = useEditorStore((state) => state.selectedPath);
  const setSelectedPath = useEditorStore((state) => state.setSelectedPath);
  const rows = [];
  walk(root, [], (node, path) => rows.push({ node, path }));

  return (
    <>
      <div className="pane-title">Tree</div>
      <div className="p-2">
        {rows.map(({ node, path }) => (
          <button
            key={path.join(".") || "root"}
            type="button"
            className={`btn btn-ghost btn-sm w-full justify-start ${samePath(path, selectedPath) ? "btn-active" : ""}`}
            style={{ paddingLeft: `${8 + path.length * 14}px` }}
            onClick={() => setSelectedPath(path)}
          >
            <span className="truncate">
              {node.type} {displayName(node)}
            </span>
          </button>
        ))}
      </div>
    </>
  );
}
