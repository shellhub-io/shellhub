import { useEffect, useRef, useState } from "react";
import { TagIcon, XMarkIcon, PlusIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useTags } from "@/hooks/useTags";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useEscapeKey } from "@/hooks/useEscapeKey";
import { useHasPermission } from "@/hooks/useHasPermission";
import { LABEL_BASE } from "@/utils/styles";

interface TagsSectionProps {
  uid: string;
  tags: string[];
  addTag: (opts: { path: { uid: string; name: string } }) => Promise<unknown>;
  removeTag: (opts: {
    path: { uid: string; name: string };
  }) => Promise<unknown>;
}

export default function TagsSection({
  uid,
  tags,
  addTag,
  removeTag,
}: TagsSectionProps) {
  const canEditTags = useHasPermission("tag:edit");
  const { tags: tagObjects } = useTags();
  const allTags = tagObjects.map((t) => t.name);
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeIndex, setActiveIndex] = useState(-1);
  const containerRef = useRef<HTMLDivElement>(null);

  const closeDropdown = () => {
    setOpen(false);
    setActiveIndex(-1);
  };

  useClickOutside(containerRef, closeDropdown);
  useEscapeKey(closeDropdown, open);

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
    if (!tag || adding) return;
    if (tags.length >= 3) return;

    const validationError = validate(tag);
    if (validationError) {
      setError(validationError);
      return;
    }
    setError(null);

    setAdding(true);
    try {
      await addTag({ path: { uid, name: tag } });
      setInput("");
      closeDropdown();
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to add tags.");
      else if (status === 400) setError(`"${tag}" is not a valid tag name.`);
      else setError("Failed to add tag.");
    }
    setAdding(false);
  };

  const handleRemove = async (tag: string) => {
    setError(null);
    try {
      await removeTag({ path: { uid, name: tag } });
    } catch (e) {
      const status = isSdkError(e) ? e.status : undefined;
      if (status === 403) setError("You don't have permission to remove tags.");
      else setError(`Failed to remove "${tag}".`);
    }
  };

  const trimmed = input.trim();
  const suggestions = allTags.filter(
    (t) => !tags.includes(t) && t.toLowerCase().includes(trimmed.toLowerCase()),
  );
  const isNew = validate(trimmed) === null && !allTags.includes(trimmed);

  const options = [
    ...suggestions.map((name) => ({
      id: `tag-opt-${name}`,
      value: name,
      isCreate: false,
    })),
    ...(isNew
      ? [{ id: "tag-opt-create", value: trimmed, isCreate: true }]
      : []),
  ];
  const showDropdown = open && trimmed && options.length > 0;
  const activeId =
    activeIndex >= 0 && activeIndex < options.length
      ? options[activeIndex].id
      : undefined;

  useEffect(() => {
    if (activeId) {
      document.getElementById(activeId)?.scrollIntoView({ block: "nearest" });
    }
  }, [activeId]);

  return (
    <div>
      <h3 className={LABEL_BASE + " mb-2"}>Tags</h3>
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
              role="combobox"
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError(null);
                setOpen(true);
                setActiveIndex(-1);
              }}
              onFocus={() => setOpen(true)}
              onKeyDown={(e) => {
                if (e.key === "ArrowDown" && showDropdown) {
                  e.preventDefault();
                  setActiveIndex((i) => (i < options.length - 1 ? i + 1 : 0));
                } else if (e.key === "ArrowUp" && showDropdown) {
                  e.preventDefault();
                  setActiveIndex((i) => (i > 0 ? i - 1 : options.length - 1));
                } else if (e.key === "Enter") {
                  e.preventDefault();
                  if (activeIndex >= 0 && activeIndex < options.length) {
                    void handleAdd(options[activeIndex].value);
                    setActiveIndex(-1);
                  } else {
                    void handleAdd(input);
                  }
                }
              }}
              placeholder="Search or create tag..."
              aria-label="New tag"
              aria-expanded={!!showDropdown}
              aria-controls={showDropdown ? "tag-suggestions" : undefined}
              aria-activedescendant={activeId}
              aria-autocomplete="list"
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

            {showDropdown && (
              <ul
                id="tag-suggestions"
                role="listbox"
                className="absolute top-full left-0 mt-1.5 z-raised w-44 max-h-[140px] overflow-y-auto bg-surface border border-border rounded-lg shadow-2xl divide-y divide-border/60"
              >
                {options.map((opt, i) => (
                  <li
                    key={opt.id}
                    id={opt.id}
                    role="option"
                    aria-selected={i === activeIndex}
                    onMouseDown={(e) => {
                      e.preventDefault();
                      void handleAdd(opt.value);
                      setActiveIndex(-1);
                    }}
                    className={`w-full text-left px-2.5 py-1.5 text-2xs ${
                      opt.isCreate ? "text-accent-green" : "text-text-primary"
                    } ${
                      i === activeIndex ? "bg-hover-medium" : ""
                    } hover:bg-hover-medium transition-colors flex items-center gap-1.5 cursor-pointer`}
                  >
                    {opt.isCreate ? (
                      <PlusIcon
                        className="w-2.5 h-2.5 shrink-0"
                        strokeWidth={2}
                      />
                    ) : (
                      <TagIcon
                        className="w-2.5 h-2.5 text-primary shrink-0"
                        strokeWidth={2}
                      />
                    )}
                    {opt.isCreate ? (
                      <>Create &ldquo;{opt.value}&rdquo;</>
                    ) : (
                      opt.value
                    )}
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </div>
      {tags.length >= 3 && (
        <p className="text-2xs text-text-muted mt-1.5">
          Maximum of 3 tags reached.
        </p>
      )}
      {error && (
        <p role="alert" className="text-2xs text-accent-red mt-1.5">
          {error}
        </p>
      )}
    </div>
  );
}
