import { useState, FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useAddMember } from "@/hooks/useMemberMutations";
import Drawer from "@/components/common/Drawer";
import { LABEL, INPUT } from "@/utils/styles";
import { RoleSelector } from "./constants";
import { type AssignableRole } from "./helpers";

const EMAIL_REGEX = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

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
  const addMember = useAddMember();
  const [email, setEmail] = useState("");
  const [role, setRole] = useState<AssignableRole>("operator");
  const [submitting, setSubmitting] = useState(false);
  const [emailError, setEmailError] = useState("");
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setEmail("");
    setRole("operator");
    setEmailError("");
    setError("");
  });

  const trimmedEmail = email.trim();
  const emailValid = EMAIL_REGEX.test(trimmedEmail);

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!emailValid) {
      setEmailError("Enter a valid email address.");
      return;
    }
    setSubmitting(true);
    setEmailError("");
    setError("");
    try {
      await addMember.mutateAsync({
        path: { tenant: tenantId },
        body: { email: trimmedEmail, role },
      });
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
      footer={(
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => void handleSubmit()}
            disabled={!emailValid || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {submitting
              ? (
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              )
              : (
                <PlusIcon className="w-4 h-4" strokeWidth={2} />
              )}
            Add Member
          </button>
        </>
      )}
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <div>
          <label className={LABEL} htmlFor="add-member-email">
            Email
          </label>
          <input
            id="add-member-email"
            type="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              if (emailError) setEmailError("");
            }}
            placeholder="user@example.com"
            autoFocus={open}
            className={`${INPUT} ${emailError ? "border-accent-red/60 focus:border-accent-red/60 focus:ring-accent-red/20" : ""}`}
            aria-invalid={!!emailError}
            aria-describedby={
              emailError ? "add-member-email-error" : undefined
            }
          />
          {emailError ? (
            <p
              id="add-member-email-error"
              className="mt-1.5 text-2xs text-accent-red"
            >
              {emailError}
            </p>
          ) : (
            <p className="text-2xs text-text-muted mt-1.5">
              Must have an existing ShellHub account
            </p>
          )}
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
