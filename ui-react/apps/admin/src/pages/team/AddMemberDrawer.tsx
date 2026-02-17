import { useState, useEffect, FormEvent } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useMembersStore } from "../../stores/membersStore";
import Drawer from "../../components/common/Drawer";
import { LABEL, INPUT } from "../../utils/styles";
import { RoleSelector } from "./constants";

/* --- Add Member Drawer --- */

function AddMemberDrawer({
  open,
  onClose,
  tenantId,
}: {
  open: boolean;
  onClose: () => void;
  tenantId: string;
}) {
  const addMember = useMembersStore((s) => s.addMember);
  const [email, setEmail] = useState("");
  const [role, setRole] = useState("operator");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setEmail("");
      setRole("operator");
      setError("");
    }
  }, [open]);

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;
    setSubmitting(true);
    setError("");
    try {
      await addMember(tenantId, email.trim(), role);
      onClose();
    } catch {
      setError("Failed to add member. Check the email and try again.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Add Member"
      footer={
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
            disabled={!email.trim() || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting ? (
              <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Add Member
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        <div>
          <label className={LABEL}>Email</label>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="user@example.com"
            autoFocus={open}
            className={INPUT}
          />
          <p className="text-2xs text-text-muted mt-1.5">
            Must have an existing ShellHub account
          </p>
        </div>
        <div>
          <label className={LABEL}>Role</label>
          <RoleSelector value={role} onChange={setRole} />
        </div>
        {error && <p className="text-2xs text-accent-red">{error}</p>}
      </form>
    </Drawer>
  );
}

export default AddMemberDrawer;
