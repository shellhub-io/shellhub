import { useState, useRef } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useTags } from "@/hooks/useTags";
import { useClickOutside } from "@/hooks/useClickOutside";
import FieldLabel from "@/components/common/fields/FieldLabel";
import FieldError from "@/components/common/fields/FieldError";
import FieldHint from "@/components/common/fields/FieldHint";

export default function TagsSelector({
  id,
  label,
  selected,
  onChange,
  error,
  hint,
}: {
  id: string;
  label: string;
  selected: string[];
  onChange: (tags: string[]) => void;
  error?: string;
  hint?: string;
}) {
  const { tags: allTags, isLoading: loading } = useTags();
  const tags = allTags.map((t) => t.name);
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const wrapperRef = useRef<HTMLDivElement>(null);
  useClickOutside(wrapperRef, () => setOpen(false));

  const filtered = tags.filter(
    (t) =>
      t.toLowerCase().includes(search.toLowerCase()) && !selected.includes(t),
  );

  const toggle = (tag: string) => {
    if (selected.includes(tag)) onChange(selected.filter((s) => s !== tag));
    else if (selected.length < 3) onChange([...selected, tag]);
  };

  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;
  const describedBy = error ? errorId : hint ? hintId : undefined;

  return (
    <div>
      <FieldLabel htmlFor={id} hideLabel>
        {label}
      </FieldLabel>
      <div ref={wrapperRef} className="relative">
        <div
          role="presentation"
          className={cn(
            "flex flex-wrap gap-1.5 min-h-[42px] px-3 py-2 bg-card border rounded-lg cursor-text transition-all",
            open ? "border-primary/50 ring-1 ring-primary/20" : "border-border",
            error && "border-accent-red/50",
          )}
          onClick={() => setOpen(true)}
        >
          {selected.map((tag) => (
            <span
              key={tag}
              className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-xs rounded-md font-medium"
            >
              {tag}
              <IconButton
                size="sm"
                aria-label="Remove tag"
                onClick={(e) => {
                  e.stopPropagation();
                  toggle(tag);
                }}
              >
                <XMarkIcon className="w-3 h-3" strokeWidth={2} />
              </IconButton>
            </span>
          ))}
          <input
            id={id}
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            onFocus={() => setOpen(true)}
            placeholder={selected.length === 0 ? "Search tags..." : ""}
            aria-invalid={error ? true : undefined}
            aria-describedby={describedBy}
            className="flex-1 min-w-[80px] bg-transparent text-sm text-text-primary placeholder:text-text-secondary outline-none"
          />
        </div>
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
      {error ? (
        <FieldError id={errorId}>{error}</FieldError>
      ) : (
        <FieldHint id={hintId}>{hint}</FieldHint>
      )}
    </div>
  );
}
