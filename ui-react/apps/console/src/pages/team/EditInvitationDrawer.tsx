import { useState } from "react";
import { InformationCircleIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateMembershipInvitation } from "@/hooks/useInvitationMutations";
import type { MembershipInvitation } from "@/client";
import Drawer from "@/components/common/Drawer";
import { LABEL } from "@/utils/styles";
import { RoleSelector } from "./constants";
import { isAssignableRole, type AssignableRole } from "./helpers";

function EditInvitationDrawer({
  open,
  onClose,
  tenantId,
  invitation,
}: {
  open: boolean;
  onClose: () => void;
  tenantId: string;
  invitation: MembershipInvitation | null;
}) {
  const updateInvitation = useUpdateMembershipInvitation();
  const [role, setRole] = useState<AssignableRole>("operator");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    // The backend may return any NamespaceMemberRole (including "owner"),
    // but only assignable roles can be chosen via RoleSelector. Fall back to
    // "operator" for any non-assignable role — in practice "owner" never
    // reaches the edit drawer, but the guard keeps types sound.
    setRole(
      isAssignableRole(invitation?.role) ? invitation.role : "operator",
    );
    setSubmitting(false);
    setError("");
  });

  const handleSubmit = async () => {
    if (!invitation) return;
    setSubmitting(true);
    setError("");
    try {
      await updateInvitation.mutateAsync({
        path: { tenant: tenantId, "user-id": invitation.user.id },
        body: { role },
      });
      onClose();
    } catch {
      setError("Failed to update invitation role. Please try again.");
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Update Invitation Role"
      subtitle={
        invitation ? (
          <span className="font-mono">{invitation.user.email}</span>
        ) : undefined
      }
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
            onClick={() => void handleSubmit()}
            disabled={role === invitation?.role || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            Save Changes
          </button>
        </>
      }
    >
      <label className={LABEL}>Role</label>
      <RoleSelector value={role} onChange={setRole} />
      {error && (
        <div
          role="alert"
          className="mt-4 flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red"
        >
          <InformationCircleIcon className="w-4 h-4 shrink-0 mt-px" />
          <span>{error}</span>
        </div>
      )}
    </Drawer>
  );
}

export default EditInvitationDrawer;
