import { useEffect, useState, useRef, useCallback } from "react";
import { createPortal } from "react-dom";
import {
  TagIcon,
  ChevronDownIcon,
  CheckIcon,
  Cog6ToothIcon,
} from "@heroicons/react/24/outline";
import { getTags } from "../../api/tags";
import { useEscapeKey } from "../../hooks/useEscapeKey";

function TagFilterDropdown({
  filterTags,
  onAdd,
  onRemove,
  onManageTags,
}: {
  filterTags: string[];
  onAdd: (tag: string) => void;
  onRemove: (tag: string) => void;
  onManageTags: () => void;
}) {
  const [open, setOpen] = useState(false);
  const [allTags, setAllTags] = useState<string[]>([]);
  const [search, setSearch] = useState("");
  const triggerRef = useRef<HTMLButtonElement>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const [pos, setPos] = useState({ top: 0, left: 0 });

  const updatePosition = useCallback(() => {
    if (!triggerRef.current) return;
    const rect = triggerRef.current.getBoundingClientRect();
    const popW = 240;
    let left = rect.left;
    if (left + popW > window.innerWidth - 12)
      left = window.innerWidth - popW - 12;
    if (left < 12) left = 12;
    setPos({ top: rect.bottom + 6, left });
  }, []);

  const [prevFilterOpen, setPrevFilterOpen] = useState(false);
  if (open && !prevFilterOpen) {
    setSearch("");
  }
  if (open !== prevFilterOpen) setPrevFilterOpen(open);

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

  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(e.target as Node) &&
        triggerRef.current &&
        !triggerRef.current.contains(e.target as Node)
      )
        setOpen(false);
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  useEscapeKey(() => setOpen(false), open);

  const filtered = allTags.filter((t) =>
    t.toLowerCase().includes(search.toLowerCase()),
  );
  const hasActive = filterTags.length > 0;

  return (
    <>
      <button
        ref={triggerRef}
        onClick={() => setOpen(!open)}
        className={`flex items-center gap-1.5 h-8 px-3 text-xs font-medium rounded-md border transition-all duration-150 ${
          hasActive
            ? "bg-primary/15 text-primary border-primary/25"
            : "bg-card text-text-muted border-border hover:text-text-secondary hover:border-border"
        }`}
      >
        <TagIcon className="w-3 h-3" strokeWidth={2} />
        Tags
        {hasActive && (
          <span className="w-4 h-4 rounded-full bg-primary text-white text-3xs font-bold flex items-center justify-center leading-none">
            {filterTags.length}
          </span>
        )}
        <ChevronDownIcon
          className={`w-3 h-3 transition-transform ${open ? "rotate-180" : ""}`}
          strokeWidth={2}
        />
      </button>

      {open &&
        createPortal(
          <div
            ref={popoverRef}
            className="fixed z-50 w-[240px] bg-surface border border-border rounded-xl shadow-2xl animate-fade-in"
            style={{ top: pos.top, left: pos.left }}
          >
            {/* Search */}
            <div className="p-2 border-b border-border">
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search tags..."
                autoFocus
                className="w-full px-2.5 py-1.5 bg-card border border-border rounded-lg text-xs text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
              />
            </div>

            {/* Tag list */}
            <div className="max-h-[200px] overflow-y-auto p-1">
              {filtered.length === 0 ? (
                <p className="px-2.5 py-3 text-2xs text-text-muted text-center">
                  No tags found
                </p>
              ) : (
                filtered.map((tag) => {
                  const active = filterTags.includes(tag);
                  return (
                    <button
                      key={tag}
                      onClick={() => {
                        if (active) onRemove(tag);
                        else onAdd(tag);
                      }}
                      className="w-full flex items-center gap-2 px-2.5 py-1.5 text-xs rounded-md hover:bg-hover-medium transition-colors"
                    >
                      <span
                        className={`w-3.5 h-3.5 rounded border flex items-center justify-center shrink-0 transition-all ${
                          active
                            ? "bg-primary border-primary"
                            : "border-text-muted/30"
                        }`}
                      >
                        {active && (
                          <CheckIcon
                            className="w-2.5 h-2.5 text-white"
                            strokeWidth={3}
                          />
                        )}
                      </span>
                      <span
                        className={`truncate ${active ? "text-primary font-medium" : "text-text-secondary"}`}
                      >
                        {tag}
                      </span>
                    </button>
                  );
                })
              )}
            </div>

            {/* Footer */}
            <div className="p-2 border-t border-border flex items-center justify-between">
              <button
                onClick={() => {
                  setOpen(false);
                  onManageTags();
                }}
                className="text-2xs text-primary hover:text-primary/80 transition-colors font-medium flex items-center gap-1"
              >
                <Cog6ToothIcon className="w-3 h-3" strokeWidth={2} />
                Manage tags
              </button>
              {hasActive && (
                <button
                  onClick={() => {
                    filterTags.forEach(onRemove);
                    setOpen(false);
                  }}
                  className="text-2xs text-text-muted hover:text-text-primary transition-colors"
                >
                  Clear all
                </button>
              )}
            </div>
          </div>,
          document.body,
        )}
    </>
  );
}

export default TagFilterDropdown;
