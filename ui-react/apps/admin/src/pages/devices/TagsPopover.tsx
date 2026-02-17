import { useEffect, useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import {
  TagIcon,
  XMarkIcon,
  PlusIcon,
  PencilIcon,
} from "@heroicons/react/24/outline";
import { addDeviceTag, removeDeviceTag } from "../../api/devices";
import { getTags, createTag } from "../../api/tags";
import { Device } from "../../types/device";
import { useEscapeKey } from "../../hooks/useEscapeKey";

/* ─── Tags Popover (portal-based, no table layout bugs) ─── */
function TagsPopover({
  device,
  onUpdated,
  onFilterTag,
}: {
  device: Device;
  onUpdated: () => void;
  onFilterTag: (tag: string) => void;
}) {
  const [open, setOpen] = useState(false);
  const [allTags, setAllTags] = useState<string[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, left: 0 });
  const tags = device.tags || [];

  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const popW = 300;
    let left = rect.left;
    if (left + popW > window.innerWidth - 12)
      left = window.innerWidth - popW - 12;
    if (left < 12) left = 12;
    setPos({ top: rect.bottom + 6, left });
  }, []);

  const [prevTagsOpen, setPrevTagsOpen] = useState(false);
  if (open && !prevTagsOpen) {
    setInput("");
    setError(null);
  }
  if (open !== prevTagsOpen) setPrevTagsOpen(open);

  useEffect(() => {
    if (!open) return;
    updatePosition();
    getTags(1, 100)
      .then(({ data }) => setAllTags(data.map((t) => t.name)))
      .catch(() => {});

    const onScroll = () => updatePosition();
    const onResize = () => updatePosition();
    window.addEventListener("scroll", onScroll, true);
    window.addEventListener("resize", onResize);
    return () => {
      window.removeEventListener("scroll", onScroll, true);
      window.removeEventListener("resize", onResize);
    };
  }, [open, updatePosition]);

  // Close on click outside
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  useEscapeKey(() => setOpen(false), open);

  const handleAdd = async (tag: string) => {
    if (tags.includes(tag) || tags.length >= 3) return;
    setLoading(true);
    setError(null);
    try {
      if (!allTags.includes(tag)) {
        await createTag(tag);
        setAllTags((prev) => [...prev, tag]);
      }
      await addDeviceTag(device.uid, tag);
      onUpdated();
      setInput("");
    } catch {
      setError(`Failed to add "${tag}"`);
    }
    setLoading(false);
  };

  const handleRemove = async (tag: string) => {
    setLoading(true);
    setError(null);
    try {
      await removeDeviceTag(device.uid, tag);
      onUpdated();
    } catch {
      setError(`Failed to remove "${tag}"`);
    }
    setLoading(false);
  };

  const suggestions = allTags.filter(
    (t) => !tags.includes(t) && t.toLowerCase().includes(input.toLowerCase()),
  );
  const isNew =
    input.trim().length >= 3 &&
    !allTags.includes(input.trim()) &&
    !tags.includes(input.trim());
  const inputValid = !input.trim() || /^[a-zA-Z0-9\-_]+$/.test(input.trim());

  return (
    <>
      {/* Trigger — inline in the table cell */}
      <div className="flex items-center gap-1 min-h-[28px] group/tags">
        {tags.length > 0 ? (
          <div className="flex items-center gap-1">
            {tags.map((tag) => (
              <button
                key={tag}
                onClick={(e) => {
                  e.stopPropagation();
                  onFilterTag(tag);
                }}
                title={`Filter by "${tag}"`}
                className="inline-flex items-center gap-0.5 px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-medium hover:bg-primary/20 transition-all cursor-pointer"
              >
                <TagIcon className="w-2 h-2" strokeWidth={2} />
                {tag}
              </button>
            ))}
          </div>
        ) : (
          <span className="text-2xs text-text-muted/30 group-hover/tags:text-text-muted transition-colors">
            No tags
          </span>
        )}
        <button
          ref={triggerRef}
          onClick={(e) => {
            e.stopPropagation();
            setOpen(!open);
          }}
          className="p-0.5 rounded text-text-muted/20 group-hover/tags:text-text-muted hover:!text-primary hover:bg-primary/10 transition-all shrink-0"
          title="Manage tags"
        >
          <PencilIcon className="w-3 h-3" strokeWidth={2} />
        </button>
      </div>

      {/* Popover — portaled to body */}
      {open &&
        createPortal(
          <div
            ref={popoverRef}
            className="fixed z-50 w-[300px] bg-surface border border-border rounded-xl shadow-2xl animate-fade-in"
            style={{ top: pos.top, left: pos.left }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="p-3 space-y-3">
              {/* Current tags */}
              {tags.length > 0 && (
                <div className="flex flex-wrap gap-1.5">
                  {tags.map((tag) => (
                    <span
                      key={tag}
                      className="inline-flex items-center gap-1 px-2 py-0.5 bg-primary/10 text-primary text-2xs rounded-md font-medium"
                    >
                      <TagIcon className="w-2.5 h-2.5" strokeWidth={2} />
                      {tag}
                      <button
                        onClick={() => handleRemove(tag)}
                        disabled={loading}
                        className="hover:text-white transition-colors disabled:opacity-dim ml-0.5"
                      >
                        <XMarkIcon className="w-2.5 h-2.5" strokeWidth={2} />
                      </button>
                    </span>
                  ))}
                </div>
              )}

              {/* Input */}
              {tags.length < 3 ? (
                <div>
                  <input
                    type="text"
                    value={input}
                    onChange={(e) => setInput(e.target.value)}
                    onKeyDown={(e) => {
                      if (
                        e.key === "Enter" &&
                        input.trim().length >= 3 &&
                        inputValid
                      ) {
                        e.preventDefault();
                        handleAdd(input.trim());
                      }
                    }}
                    placeholder="Search or create tag..."
                    autoFocus
                    className="w-full px-2.5 py-1.5 bg-card border border-border rounded-lg text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
                  />
                  {input.trim() && input.trim().length < 3 && (
                    <p className="text-2xs text-text-muted mt-1">
                      Min 3 characters
                    </p>
                  )}
                  {input.trim().length >= 3 && !inputValid && (
                    <p className="text-2xs text-accent-red mt-1">
                      Only letters, numbers, - and _
                    </p>
                  )}

                  {/* Suggestions dropdown */}
                  {(suggestions.length > 0 || isNew) &&
                    input.trim() &&
                    inputValid && (
                      <div className="mt-1.5 max-h-[140px] overflow-y-auto border border-border rounded-lg divide-y divide-border/60">
                        {suggestions.map((tag) => (
                          <button
                            key={tag}
                            onClick={() => handleAdd(tag)}
                            disabled={loading}
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
                            onClick={() => handleAdd(input.trim())}
                            disabled={loading}
                            className="w-full text-left px-2.5 py-1.5 text-2xs text-accent-green hover:bg-hover-medium transition-colors disabled:opacity-dim flex items-center gap-1.5"
                          >
                            <PlusIcon
                              className="w-2.5 h-2.5 shrink-0"
                              strokeWidth={2}
                            />
                            Create &ldquo;{input.trim()}&rdquo;
                          </button>
                        )}
                      </div>
                    )}
                </div>
              ) : (
                <p className="text-2xs text-text-muted">
                  Max 3 tags. Remove one to add another.
                </p>
              )}

              {error && (
                <p className="text-2xs font-mono text-accent-red">{error}</p>
              )}
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}

export default TagsPopover;
