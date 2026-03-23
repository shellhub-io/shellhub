import { useState, useRef } from "react";
import { ChevronDownIcon, PlusIcon } from "@heroicons/react/24/outline";
import { useNamespaces, useNamespace } from "../../hooks/useNamespaces";
import { useSwitchNamespace } from "../../hooks/useNamespaceMutations";
import { useAuthStore } from "../../stores/authStore";
import { useClickOutside } from "../../hooks/useClickOutside";
import CreateNamespaceDialog from "../common/CreateNamespaceDialog";

export default function NamespaceSelector() {
  const { namespaces } = useNamespaces();
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const switchNs = useSwitchNamespace();

  const [open, setOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useClickOutside(containerRef, () => setOpen(false));

  const others = namespaces.filter(
    (ns) => ns.tenant_id !== currentNamespace?.tenant_id,
  );

  const handleSwitch = async (id: string) => {
    setOpen(false);
    await switchNs.mutateAsync(id);
  };

  const handleCreate = () => {
    setOpen(false);
    setCreateOpen(true);
  };

  const initials = (name: string) =>
    name
      .split(/[\s-_]+/)
      .slice(0, 2)
      .map((w) => w[0]?.toUpperCase() ?? "")
      .join("");

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2.5 h-9 px-3 rounded-md border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
      >
        {currentNamespace
          ? (
            <>
              <span className="w-6 h-6 rounded bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
                {initials(currentNamespace.name)}
              </span>
              <span className="text-sm font-medium text-text-primary max-w-[180px] truncate">
                {currentNamespace.name}
              </span>
            </>
          )
          : (
            <span className="text-sm text-text-muted italic">No namespace</span>
          )}
        <ChevronDownIcon
          className={`w-3 h-3 text-text-muted transition-transform duration-200 ${open ? "rotate-180" : ""}`}
          strokeWidth={2.5}
        />
      </button>

      {open && (
        <div className="absolute top-full left-0 mt-1.5 w-80 bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down">
          {currentNamespace && (
            <div className="p-4 border-b border-border">
              <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
                Active Namespace
              </p>
              <div className="flex items-center gap-3">
                <span className="w-8 h-8 rounded-md bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold font-mono shrink-0">
                  {initials(currentNamespace.name)}
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-semibold text-text-primary truncate">
                    {currentNamespace.name}
                  </p>
                  <p className="text-2xs font-mono text-text-muted truncate mt-0.5">
                    {currentNamespace.tenant_id}
                  </p>
                </div>
                <span className="text-2xs font-mono text-accent-green bg-accent-green/10 px-2 py-0.5 rounded-full border border-accent-green/20">
                  active
                </span>
              </div>
            </div>
          )}

          {others.length > 0 && (
            <div className="p-2">
              <p className="px-2 py-1.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
                Switch Namespace
              </p>
              {others.map((ns) => (
                <button
                  key={ns.tenant_id}
                  onClick={() => void handleSwitch(ns.tenant_id)}
                  className="w-full flex items-center gap-3 px-2 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
                >
                  <span className="w-7 h-7 rounded bg-card border border-border flex items-center justify-center text-text-muted text-2xs font-bold font-mono group-hover:border-primary/30 group-hover:text-primary transition-colors shrink-0">
                    {initials(ns.name)}
                  </span>
                  <div className="min-w-0 flex-1">
                    <p className="text-sm text-text-secondary group-hover:text-text-primary truncate transition-colors">
                      {ns.name}
                    </p>
                    <p className="text-2xs font-mono text-text-muted truncate">
                      {ns.devices_accepted_count}
                      {" "}
                      device
                      {ns.devices_accepted_count !== 1 ? "s" : ""}
                    </p>
                  </div>
                </button>
              ))}
            </div>
          )}

          {others.length === 0 && !currentNamespace && (
            <div className="p-6 text-center">
              <p className="text-xs text-text-muted">No namespaces available</p>
            </div>
          )}

          {/* Create namespace */}
          <div className="p-2 border-t border-border">
            <button
              onClick={handleCreate}
              className="w-full flex items-center gap-3 px-2 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
            >
              <span className="w-7 h-7 rounded bg-card border border-dashed border-border-light flex items-center justify-center text-text-muted group-hover:border-primary/40 group-hover:text-primary transition-colors shrink-0">
                <PlusIcon className="w-3.5 h-3.5" strokeWidth={2} />
              </span>
              <p className="text-sm text-text-muted group-hover:text-text-primary transition-colors">
                Create namespace
              </p>
            </button>
          </div>
        </div>
      )}

      <CreateNamespaceDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
    </div>
  );
}
