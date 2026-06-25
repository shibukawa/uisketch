export function SourceEditor({ yaml }) {
  return (
    <pre data-testid="yaml" id="yaml" className="min-h-0 flex-1 overflow-auto bg-neutral p-4 font-mono text-sm leading-relaxed text-neutral-content">
      {yaml}
    </pre>
  );
}
