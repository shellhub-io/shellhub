import { useState, useEffect, FormEvent } from "react";
import { KeyIcon, CheckIcon } from "@heroicons/react/24/outline";
import { useApiKeysStore } from "../../stores/apiKeysStore";
import CopyButton from "../../components/common/CopyButton";
import Drawer from "../../components/common/Drawer";
import { LABEL, INPUT } from "../../utils/styles";
import { RoleSelector } from "./constants";
import { EXPIRY_OPTIONS } from "./helpers";

/* --- Generate API Key Drawer --- */

function GenerateKeyDrawer({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const generate = useApiKeysStore((s) => s.generate);
  const [name, setName] = useState("");
  const [role, setRole] = useState("administrator");
  const [expiresIn, setExpiresIn] = useState(30);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [generatedKey, setGeneratedKey] = useState("");

  useEffect(() => {
    if (open) {
      setName("");
      setRole("administrator");
      setExpiresIn(30);
      setError("");
      setGeneratedKey("");
    }
  }, [open]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;
    setSubmitting(true);
    setError("");
    try {
      const id = await generate(name.trim(), role, expiresIn);
      setGeneratedKey(id);
    } catch {
      setError("Failed to generate API key. The name may already exist.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Generate API Key"
      footer={
        generatedKey ? (
          <button
            onClick={onClose}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
          >
            Done
          </button>
        ) : (
          <>
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSubmit}
              disabled={!name.trim() || submitting}
              className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
            >
              {submitting ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <KeyIcon className="w-4 h-4" strokeWidth={2} />
              )}
              Generate Key
            </button>
          </>
        )
      }
    >
      {generatedKey ? (
        <div className="space-y-5">
          <div className="flex items-start gap-3 bg-accent-green/[0.06] border border-accent-green/20 rounded-xl px-4 py-3.5">
            <CheckIcon className="w-5 h-5 text-accent-green shrink-0 mt-0.5" />
            <div>
              <p className="text-sm font-medium text-text-primary">
                API Key Generated
              </p>
              <p className="text-2xs text-text-muted mt-0.5">
                Copy this key now. You won't be able to see it again.
              </p>
            </div>
          </div>
          <div>
            <label className={LABEL}>Your API Key</label>
            <div className="flex items-center gap-2 bg-card border border-border rounded-lg px-3.5 py-2.5">
              <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
                {generatedKey}
              </code>
              <CopyButton text={generatedKey} size="md" />
            </div>
          </div>
        </div>
      ) : (
        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className={LABEL}>Name</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. ci-pipeline"
              autoFocus={open}
              className={INPUT}
            />
          </div>
          <div>
            <label className={LABEL}>Role</label>
            <RoleSelector value={role} onChange={setRole} />
          </div>
          <div>
            <label className={LABEL}>Expiration</label>
            <div className="flex flex-wrap gap-1.5">
              {EXPIRY_OPTIONS.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => setExpiresIn(opt.value)}
                  className={`px-3 py-1.5 text-xs font-medium rounded-md border transition-all ${
                    expiresIn === opt.value
                      ? "bg-primary/[0.08] border-primary/30 text-primary ring-1 ring-primary/10"
                      : "bg-card border-border text-text-secondary hover:border-border-light"
                  }`}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
          {error && <p className="text-2xs text-accent-red">{error}</p>}
        </form>
      )}
    </Drawer>
  );
}

export default GenerateKeyDrawer;
