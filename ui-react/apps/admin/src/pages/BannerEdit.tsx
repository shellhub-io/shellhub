import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { CheckIcon, ArrowLeftIcon } from "@heroicons/react/24/outline";
import { useNamespacesStore } from "../stores/namespacesStore";
import { useAuthStore } from "../stores/authStore";
import { Namespace } from "../types/namespace";

const MAX_LENGTH = 4096;

/* --- Editor (mounts only after namespace is loaded) --- */

function BannerEditor({ ns, canEdit }: { ns: Namespace; canEdit: boolean }) {
  const updateNamespace = useNamespacesStore((s) => s.updateNamespace);
  const navigate = useNavigate();

  const banner = ns.settings?.connection_announcement ?? "";
  const [text, setText] = useState(banner);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const changed = text !== banner;
  const overLimit = text.length > MAX_LENGTH;

  const handleSave = async () => {
    if (!changed || overLimit || saving) return;
    setSaving(true);
    setError("");
    try {
      await updateNamespace(ns.tenant_id, {
        settings: { connection_announcement: text },
      });
      navigate("/settings");
    } catch {
      setError("Failed to save. Please try again.");
      setSaving(false);
    }
  };

  return (
    <div className="flex flex-col flex-1 min-h-0 animate-fade-in">
      {/* Textarea wrapper -- grows to fill */}
      <div
        className={`flex flex-col flex-1 min-h-0 rounded-xl border overflow-hidden transition-colors ${
          overLimit
            ? "border-accent-red/30"
            : changed
              ? "border-primary/30"
              : "border-border"
        }`}
      >
        <textarea
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            setError("");
          }}
          disabled={!canEdit}
          placeholder="Enter the banner message..."
          className="flex-1 w-full px-4 py-3.5 bg-card text-sm text-text-primary font-mono placeholder:text-text-muted/30 focus:outline-none transition-all resize-none leading-relaxed disabled:opacity-dim disabled:cursor-not-allowed"
          autoFocus
        />
        <div className="flex items-center justify-between px-4 py-2.5 border-t border-border bg-surface/50 shrink-0">
          <span
            className={`text-2xs font-mono ${overLimit ? "text-accent-red font-semibold" : "text-text-muted/50"}`}
          >
            {text.length.toLocaleString()}/{MAX_LENGTH.toLocaleString()}
          </span>
          {overLimit && (
            <span className="text-2xs text-accent-red">
              Exceeds maximum length
            </span>
          )}
        </div>
      </div>

      {/* Footer */}
      <div className="shrink-0">
        <p className="text-2xs text-text-muted mt-3 leading-relaxed">
          Supports plain text only. Common uses: legal notices, maintenance
          alerts, and security warnings.
        </p>

        {error && (
          <div className="mt-4 px-3.5 py-2.5 rounded-lg bg-accent-red/10 border border-accent-red/20 text-sm text-accent-red">
            {error}
          </div>
        )}

        {canEdit && (
          <div className="flex items-center justify-end gap-2 mt-6 pt-6 border-t border-border">
            <Link
              to="/settings"
              className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </Link>
            <button
              onClick={handleSave}
              disabled={!changed || overLimit || saving}
              className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {saving ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <CheckIcon className="w-4 h-4" strokeWidth={2} />
              )}
              Save
            </button>
          </div>
        )}
      </div>
    </div>
  );
}

/* --- Page --- */

export default function BannerEdit() {
  const { userId, tenant: tenantId, role: sessionRole } = useAuthStore();
  const { currentNamespace: ns, fetchCurrent } = useNamespacesStore();

  useEffect(() => {
    if (tenantId && !ns) fetchCurrent(tenantId);
  }, [tenantId, ns, fetchCurrent]);

  if (!ns) {
    return (
      <div className="flex items-center justify-center py-32">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
      </div>
    );
  }

  const isOwner = ns.owner === userId;
  const currentMember = ns.members?.find((m) => m.id === userId);
  const role =
    currentMember?.role ?? (isOwner ? "owner" : (sessionRole ?? "observer"));
  const canEdit = isOwner || role === "administrator";

  return (
    <div className="flex flex-col flex-1 min-h-0">
      {/* Header */}
      <div className="relative -mx-8 -mt-8 px-8 py-6 mb-8 border-b border-border bg-surface shrink-0">
        <div className="flex items-start gap-4">
          <Link
            to="/settings"
            className="w-12 h-12 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted hover:text-text-primary hover:border-border-light transition-all shrink-0"
          >
            <ArrowLeftIcon className="w-5 h-5" />
          </Link>
          <div>
            <p className="text-2xs font-mono font-semibold uppercase tracking-label text-primary mb-1">
              Settings
            </p>
            <h1 className="text-xl font-semibold text-text-primary leading-tight">
              SSH Banner
            </h1>
            <p className="text-sm text-text-muted mt-1 max-w-xl">
              This message is displayed when users connect to any device in this
              namespace via SSH.
            </p>
          </div>
        </div>
      </div>

      <BannerEditor ns={ns} canEdit={canEdit} />
    </div>
  );
}
