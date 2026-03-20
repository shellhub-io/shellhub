import { useState, useRef } from "react";
import {
  ChevronDownIcon,
  PlusIcon,
  ShieldCheckIcon,
} from "@heroicons/react/24/outline";
import { useNamespaces, useNamespace } from "../../hooks/useNamespaces";
import { useSwitchNamespace } from "../../hooks/useNamespaceMutations";
import { useAuthStore } from "../../stores/authStore";
import { useClickOutside } from "../../hooks/useClickOutside";
import { getInitials } from "../../utils/string";
import { getConfig } from "@/env";
import CreateNamespaceDialog from "../common/CreateNamespaceDialog";
import { useNavigate } from "react-router-dom";

const ADMIN_SUBTITLE = "Super Admin \u00B7 Instance";

interface NamespaceSelectorProps {
  isAdminContext?: boolean;
}

export default function NamespaceSelector({
  isAdminContext = false,
}: NamespaceSelectorProps) {
  const { namespaces } = useNamespaces();
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const { namespace: currentNamespace } = useNamespace(tenantId);
  const switchNs = useSwitchNamespace();
  const navigate = useNavigate();

  const [open, setOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useClickOutside(containerRef, () => setOpen(false));

  const showAdminLink
    = !isAdminContext
      && getConfig().enterprise
      && !getConfig().cloud
      && isAdmin;

  const availableNamespaces = isAdminContext
    ? namespaces
    : namespaces.filter(
      (ns) => ns.tenant_id !== currentNamespace?.tenant_id,
    );

  const handleSwitch = async (id: string) => {
    setOpen(false);
    await switchNs.mutateAsync({
      tenantId: id,
      redirectTo: isAdminContext ? "/dashboard" : undefined,
    });
  };

  const handleCreate = () => {
    setOpen(false);
    setCreateOpen(true);
  };

  return (
    <div ref={containerRef} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className="flex items-center gap-2.5 h-9 px-3 rounded-md border border-transparent hover:border-border hover:bg-hover-subtle transition-all duration-150"
      >
        {isAdminContext
          ? (
            <>
              <ShieldCheckIcon className="w-5 h-5 text-accent-red" />
              <span className="hidden md:inline text-sm font-medium text-text-primary">
                Admin Console
              </span>
            </>
          )
          : currentNamespace
            ? (
              <>
                <span className="w-6 h-6 rounded bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono">
                  {getInitials(currentNamespace.name)}
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
        <div className="absolute top-full left-0 mt-1.5 w-80 max-w-[calc(100vw-2rem)] bg-surface border border-border rounded-lg shadow-2xl shadow-black/40 z-50 overflow-hidden animate-slide-down">
          {/* Active namespace header (non-admin only) */}
          {!isAdminContext && currentNamespace && (
            <div className="p-4 border-b border-border">
              <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
                Active Namespace
              </p>
              <div className="flex items-center gap-3">
                <span className="w-8 h-8 rounded-md bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-xs font-bold font-mono shrink-0">
                  {getInitials(currentNamespace.name)}
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

          {/* Admin console header (admin context only) */}
          {isAdminContext && (
            <div className="p-4 border-b border-border">
              <div className="flex items-center gap-3">
                <span className="w-8 h-8 rounded-md bg-accent-red/15 border border-accent-red/20 flex items-center justify-center shrink-0">
                  <ShieldCheckIcon className="w-4 h-4 text-accent-red" />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm font-semibold text-text-primary">
                    Admin Console
                  </p>
                  <p className="text-2xs font-mono text-text-muted mt-0.5">
                    {ADMIN_SUBTITLE}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Namespace list */}
          {availableNamespaces.length > 0 && (
            <div className="p-2">
              <p className="px-2 py-1.5 text-2xs font-mono font-semibold uppercase tracking-label text-text-muted">
                {isAdminContext ? "Available Namespaces" : "Switch Namespace"}
              </p>
              {availableNamespaces.map((ns) => (
                <button
                  key={ns.tenant_id}
                  onClick={() => void handleSwitch(ns.tenant_id)}
                  className="w-full flex items-center gap-3 px-2 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
                >
                  <span className="w-7 h-7 rounded bg-card border border-border flex items-center justify-center text-text-muted text-2xs font-bold font-mono group-hover:border-primary/30 group-hover:text-primary transition-colors shrink-0">
                    {getInitials(ns.name)}
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

          {availableNamespaces.length === 0 && !isAdminContext && !currentNamespace && (
            <div className="p-6 text-center">
              <p className="text-xs text-text-muted">No namespaces available</p>
            </div>
          )}

          {/* Admin Console link (non-admin context, enterprise admins) */}
          {showAdminLink && (
            <div className="p-2 border-t border-border">
              <button
                onClick={() => {
                  setOpen(false);
                  void navigate("/admin");
                }}
                className="w-full flex items-center gap-3 px-2 py-2 rounded-md text-left hover:bg-hover-medium transition-colors group"
              >
                <span className="w-7 h-7 rounded bg-accent-red/10 border border-accent-red/20 flex items-center justify-center text-accent-red group-hover:bg-accent-red/15 transition-colors shrink-0">
                  <ShieldCheckIcon className="w-3.5 h-3.5" strokeWidth={2} />
                </span>
                <div className="min-w-0 flex-1">
                  <p className="text-sm text-text-secondary group-hover:text-text-primary transition-colors">
                    Admin Console
                  </p>
                  <p className="text-2xs font-mono text-text-muted">
                    {ADMIN_SUBTITLE}
                  </p>
                </div>
              </button>
            </div>
          )}

          {/* Create namespace (non-admin context only) */}
          {!isAdminContext && (
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
          )}
        </div>
      )}

      <CreateNamespaceDialog
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
    </div>
  );
}
