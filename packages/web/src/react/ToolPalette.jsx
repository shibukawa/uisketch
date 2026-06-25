import { paletteGroups } from "../model/component-catalog.js";
import { useEditorStore } from "../state/useEditorStore.js";

export function ToolPalette() {
  const selectedPath = useEditorStore((state) => state.selectedPath);
  const insertNewNode = useEditorStore((state) => state.insertNewNode);
  const setDragging = useEditorStore((state) => state.setDragging);

  return (
    <>
      <div className="pane-title">Components</div>
      <div className="grid gap-4 p-3">
        {paletteGroups.map((group) => (
          <section key={group.id} className="grid gap-2">
            <h3 className="palette-group-title">{group.label}</h3>
            <div className="grid grid-cols-2 gap-1">
              {group.items.map((item) => (
                <button
                  key={item.type}
                  type="button"
                  draggable
                  className="palette-item"
                  onClick={() => insertNewNode(item.type, selectedPath)}
                  onDragStart={(event) => {
                    setDragging({ kind: "new", type: item.type });
                    event.dataTransfer.effectAllowed = "copy";
                    event.dataTransfer.setData("text/plain", item.type);
                  }}
                  onDragEnd={() => setDragging(null)}
                >
                  <span className={`palette-icon palette-icon-${item.type}`} aria-hidden="true">
                    <span />
                  </span>
                  <span className="truncate">{item.label}</span>
                </button>
              ))}
            </div>
          </section>
        ))}
      </div>
    </>
  );
}
