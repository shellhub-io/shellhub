import { useEffect, useState, useRef } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { getTags } from "../../api/tags";
import { useClickOutside } from "../../hooks/useClickOutside";

export default function TagsSelector({
  selected,
  onChange,
  error,
}: {
  selected: string[];
  onChange: (tags: string[]) => void;
  error?: string;
}) {
  const [tags, setTags] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    getTags(1, 100)
      .then(({ data }) => setTags(data.map((t) => t.name)))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  useClickOutside(wrapperRef, () => setOpen(false));

  const filtered = tags.filter(
    (t) =>
      t.toLowerCase().includes(search.toLowerCase()) && !selected.includes(t),
  );

  const toggle = (tag: string) => {
    if (selected.includes(tag)) onChange(selected.filter((s) => s !== tag));
    else if (selected.length < 3) onChange([...selected, tag]);
  };

  return (
    <div ref={wrapperRef} className="relative">
      <div
        className={`flex flex-wrap gap-1.5 min-h-[42px] px-3 py-2 bg-card border rounded-lg cursor-text transition-all ${
          open ? "border-primary/50 ring-1 ring-primary/20" : "border-border"
        } ${error ? "border-accent-red/50" : ""}`}
        onClick={() => setOpen(true)}
      >
        {selected.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-md font-medium"
          >
            {tag}
            <button
              type="button"
              onClick={(e) => {
                e.stopPropagation();
                toggle(tag);
              }}
              className="hover:text-white transition-colors"
            >
              <XMarkIcon className="w-3 h-3" strokeWidth={2} />
            </button>
          </span>
        ))}
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onFocus={() => setOpen(true)}
          placeholder={selected.length === 0 ? "Search tags..." : ""}
          className="flex-1 min-w-[80px] bg-transparent text-sm text-text-primary placeholder:text-text-secondary outline-none"
        />
      </div>
      {error && <p className="mt-1 text-2xs text-accent-red">{error}</p>}
      {open && (
        <div className="absolute z-10 mt-1 w-full max-h-48 overflow-y-auto bg-surface border border-border rounded-lg shadow-xl">
          {loading ? (
            <div className="px-3 py-2 text-xs text-text-muted">
              Loading tags...
            </div>
          ) : filtered.length === 0 ? (
            <div className="px-3 py-2 text-xs text-text-muted">
              No tags found
            </div>
          ) : (
            filtered.map((tag) => (
              <button
                key={tag}
                type="button"
                onClick={() => {
                  toggle(tag);
                  setSearch("");
                }}
                className="w-full text-left px-3 py-2 text-sm text-text-primary hover:bg-hover-medium transition-colors"
              >
                {tag}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
