import { useRef, useState } from "react";
import { TagIcon, XMarkIcon, PlusIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import {
  useAddDeviceTag,
  useRemoveDeviceTag,
} from "@/hooks/useDeviceMutations";
import { useTags } from "@/hooks/useTags";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { useHasPermission } from "@/hooks/useHasPermission";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

interface TagsSectionProps {
  uid: string;
  tags: string[];
}

export default function TagsSection({ uid, tags }: TagsSectionProps) {
  const addTagMutation = useAddDeviceTag();
  const removeTagMutation = useRemoveDeviceTag();
  const canEditTags = useHasPermission("tag:edit");
  const { tags: tagObjects } = useTags();
  const allTags = tagObjects.map((t) => t.name);
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useClickOutside(containerRef, () => setOpen(false));
  useEscapeKey(() => setOpen(false), open);

  const validate = (tag: string): string | null => {
    if (tags.includes(tag)) return "This tag is already added.";
    if (tag.length < 3) return "Tag must be at least 3 characters.";
    if (tag.length > 255) return "Tag must be at most 255 characters.";
    if (!/^[a-zA-Z0-9]+$/.test(tag))
      return "Tag must contain only letters and numbers.";
    return null;
  };

  const handleAdd = async (value: string) => {
    const tag = value.trim();
    if (!tag) return;
    if (tags.length >= 3) return;

    const validationError = validate(tag);
    if (validationError) {
      setError(validationError);
      return;
    }
    setError(null);

    setAdding(true);
    try {
      await addTagMutation.mutateAsync({ path: { uid, name: tag } });
      setInput("");
      setOpen(false);
    } catch {
      setError("Failed to add tag.");
    }
    setAdding(false);
  };

  const handleRemove = async (tag: string) => {
    try {
      await removeTagMutation.mutateAsync({ path: { uid, name: tag } });
    } catch {
      /* invalidation handles UI update */
    }
  };

  const trimmed = input.trim();
  const suggestions = allTags.filter(
    (t) => !tags.includes(t) && t.toLowerCase().includes(trimmed.toLowerCase()),
  );
  const isNew = validate(trimmed) === null && !allTags.includes(trimmed);

  return (
    <div>
      <h3 className={LABEL + " mb-2"}>Tags</h3>
      <div className="flex flex-wrap items-center gap-2">
        {tags.map((tag) => (
          <span
            key={tag}
            className="inline-flex items-center gap-1.5 px-2.5 py-1 bg-primary/10 text-primary text-xs rounded-md font-medium"
          >
            <TagIcon className="w-3 h-3" strokeWidth={2} />
            {tag}
            {canEditTags && (
              <button
                type="button"
                onClick={() => void handleRemove(tag)}
                aria-label={`Remove tag ${tag}`}
                className="hover:text-white transition-colors"
              >
                <XMarkIcon className="w-3 h-3" strokeWidth={2} />
              </button>
            )}
          </span>
        ))}
        {canEditTags && tags.length < 3 && (
          <div
            ref={containerRef}
            className="relative flex items-center gap-1.5"
          >
            <input
              type="text"
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError(null);
                setOpen(true);
              }}
              onFocus={() => setOpen(true)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  e.preventDefault();
                  void handleAdd(input);
                }
              }}
              placeholder="Search or create tag..."
              aria-label="New tag"
              className="w-44 px-2.5 py-1 bg-card border border-border rounded-md text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/40 transition-all"
            />
            <IconButton
              variant="primary"
              size="sm"
              disabled={adding || !trimmed}
              aria-label="Add tag"
              onClick={() => void handleAdd(input)}
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            </IconButton>

            {open && trimmed && (suggestions.length > 0 || isNew) && (
              <div className="absolute top-full left-0 mt-1.5 z-10 w-44 max-h-[140px] overflow-y-auto bg-surface border border-border rounded-lg shadow-2xl divide-y divide-border/60">
                {suggestions.map((tag) => (
                  <button
                    type="button"
                    key={tag}
                    onClick={() => void handleAdd(tag)}
                    disabled={adding}
                    className="w-full text-left px-2.5 py-1.5 text-2xs text-text-primary hover:bg-hover-medium transition-colors disabled:opacity-dim flex items-center gap-1.5"
                  >
                    <TagIcon
                      className="w-2.5 h-2.5 text-primary shrink-0"
                      strokeWidth={2}
                    />
                    {tag}
                  </button>
                ))}
                {isNew && (
                  <button
                    type="button"
                    onClick={() => void handleAdd(input)}
                    disabled={adding}
                    className="w-full text-left px-2.5 py-1.5 text-2xs text-accent-green hover:bg-hover-medium transition-colors disabled:opacity-dim flex items-center gap-1.5"
                  >
                    <PlusIcon
                      className="w-2.5 h-2.5 shrink-0"
                      strokeWidth={2}
                    />
                    Create &ldquo;
                    {trimmed}
                    &rdquo;
                  </button>
                )}
              </div>
            )}
          </div>
        )}
      </div>
      {tags.length >= 3 && (
        <p className="text-2xs text-text-muted mt-1.5">
          Maximum of 3 tags reached.
        </p>
      )}
      {error && <p className="text-2xs text-accent-red mt-1.5">{error}</p>}
    </div>
  );
}
