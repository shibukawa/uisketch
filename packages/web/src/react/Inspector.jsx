import { useEffect, useMemo, useState } from "react";
import { fieldsFor, rootSurfaceTypes } from "../model/component-catalog.js";
import { useEditorStore } from "../state/useEditorStore.js";

export function Inspector() {
  const node = useEditorStore((state) => state.selectedNode());
  const selectedPath = useEditorStore((state) => state.selectedPath);
  const updateSelectedField = useEditorStore((state) => state.updateSelectedField);
  const deleteSelected = useEditorStore((state) => state.deleteSelected);
  const replaceRootType = useEditorStore((state) => state.replaceRootType);
  const selectedPathKey = useMemo(() => selectedPath.join("."), [selectedPath]);
  const [gridColumnsDraft, setGridColumnsDraft] = useState("");
  const [tableColumnsDraft, setTableColumnsDraft] = useState("");
  const [tabsLabelsDraft, setTabsLabelsDraft] = useState("");

  useEffect(() => {
    if (node?.type === "grid") setGridColumnsDraft(String(node.columns || 2));
  }, [node?.type, node?.columns, selectedPathKey]);

  useEffect(() => {
    if (node?.type === "table") setTableColumnsDraft((node.columns || []).join(", "));
  }, [node?.type, selectedPathKey]);

  useEffect(() => {
    if (node?.type === "tabs") setTabsLabelsDraft(formatLabelsDraft(node.labels || []));
  }, [node?.type, selectedPathKey]);

  if (!node) return <div className="p-3 text-sm text-base-content/60">Select a component.</div>;

  return (
    <form className="grid gap-3 p-3" onSubmit={(event) => event.preventDefault()}>
      {selectedPath.length === 0 ? (
        <SelectField label="type" value={node.type} options={rootSurfaceTypes} onChange={replaceRootType} />
      ) : (
        <Field label="type" value={node.type} disabled />
      )}
      <TextAreaField
        label="prompt"
        value={node.prompt || ""}
        help="AI-facing note saved with the node."
        onChange={(value) => updateSelectedField("prompt", value)}
      />
      <TextAreaField
        label="data"
        value={node.data || ""}
        help="Freeform node metadata saved with the node."
        onChange={(value) => updateSelectedField("data", value)}
      />
      {fieldsFor(node).map((key) => (
        <Field key={key} label={key} value={node[key] || ""} onChange={(value) => updateSelectedField(key, value)} />
      ))}
      {node.type === "grid" && (
        <Field
          label="columns"
          value={gridColumnsDraft}
          invalid={!isValidPositiveInteger(gridColumnsDraft)}
          help={!isValidPositiveInteger(gridColumnsDraft) ? `Using last valid value: ${node.columns || 2}` : ""}
          onChange={(value) => {
            setGridColumnsDraft(value);
            if (isValidPositiveInteger(value)) updateSelectedField("columns", Number(value));
          }}
        />
      )}
      {node.type === "table" && (
        <Field
          label="columns"
          value={tableColumnsDraft}
          onChange={(value) => {
            setTableColumnsDraft(value);
            updateSelectedField("columns", splitCommaList(value));
          }}
        />
      )}
      {node.type === "select" && (
        <Field
          label="options"
          value={(node.options || []).join(", ")}
          onChange={(value) => updateSelectedField("options", splitCommaList(value))}
        />
      )}
      {node.type === "tabs" && (
        <>
          <Field
            label="tabs"
            value={tabsLabelsDraft}
            onChange={(value) => {
              setTabsLabelsDraft(value);
              updateSelectedField("labels", parseLabelsDraft(value, selectedLabelIndex(node.labels || [])));
            }}
          />
          <SelectField
            label="active tab"
            value={String(selectedLabelIndex(node.labels || []))}
            options={labelOptions(node.labels || [])}
            onChange={(value) => updateSelectedField("labels", selectLabelIndex(node.labels || [], Number(value)))}
          />
        </>
      )}
      {selectedPath.length > 0 && (
        <button type="button" className="btn btn-outline btn-error" onClick={deleteSelected}>
          Delete
        </button>
      )}
    </form>
  );
}

function SelectField({ label, value, options, onChange }) {
  return (
    <label className="form-control">
      <span className="label pb-1 text-xs text-base-content/60">{label}</span>
      <select className="select select-bordered select-sm w-full" value={value} onChange={(event) => onChange?.(event.target.value)}>
        {options.map((option) => {
          const optionValue = typeof option === "string" ? option : option.value;
          const optionLabel = typeof option === "string" ? option : option.label;
          return (
            <option key={optionValue} value={optionValue}>
              {optionLabel}
            </option>
          );
        })}
      </select>
    </label>
  );
}

function Field({ label, value, onChange, disabled = false, invalid = false, help = "" }) {
  return (
    <label className="form-control">
      <span className="label pb-1 text-xs text-base-content/60">{label}</span>
      <input
        className={`input input-bordered input-sm w-full ${invalid ? "input-error" : ""}`}
        value={value}
        disabled={disabled}
        onChange={(event) => onChange?.(event.target.value)}
      />
      {help && <span className="pt-1 text-xs text-error">{help}</span>}
    </label>
  );
}

function TextAreaField({ label, value, onChange, disabled = false, help = "" }) {
  return (
    <label className="form-control">
      <span className="label pb-1 text-xs text-base-content/60">{label}</span>
      <textarea
        className="textarea textarea-bordered textarea-sm min-h-24 w-full"
        value={value}
        disabled={disabled}
        onChange={(event) => onChange?.(event.target.value)}
      />
      {help && <span className="pt-1 text-xs text-base-content/50">{help}</span>}
    </label>
  );
}

function isValidPositiveInteger(value) {
  return /^[1-9]\d*$/.test(value);
}

function splitCommaList(value) {
  return value
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function formatLabelsDraft(labels) {
  return labels.map((label) => label.text || "").join(", ");
}

function parseLabelsDraft(value, selectedIndex) {
  const items = splitCommaList(value);
  const boundedSelectedIndex = Math.min(Math.max(selectedIndex, 0), Math.max(items.length - 1, 0));
  return items.map((text, index) => ({
    text,
    selected: index === boundedSelectedIndex,
  }));
}

function selectedLabelIndex(labels) {
  return Math.max(0, labels.findIndex((label) => label.selected));
}

function selectLabelIndex(labels, selectedIndex) {
  const boundedSelectedIndex = Math.min(Math.max(selectedIndex, 0), Math.max(labels.length - 1, 0));
  return labels.map((label, index) => ({
    ...label,
    selected: index === boundedSelectedIndex,
  }));
}

function labelOptions(labels) {
  const source = labels.length ? labels : [{ text: "Tab 1", selected: true }];
  return source.map((label, index) => ({
    value: String(index),
    label: label.text || `Tab ${index + 1}`,
  }));
}
