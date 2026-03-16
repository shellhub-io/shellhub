import { useState } from "react";
import { useResetOnOpen } from "../../hooks/useResetOnOpen";
import { useUpdateMemberRole } from "../../hooks/useMemberMutations";
import { type NamespaceMember } from "../../hooks/useNamespaces";
import type { NamespaceMemberRole } from "../../client";
import Drawer from "../../components/common/Drawer";
import { LABEL } from "../../utils/styles";
import { RoleSelector } from "./constants";

/* ─── Edit Member Drawer ─── */

function EditMemberDrawer({
  open,
  onClose,
  tenantId,
  member,
}: {
  open: boolean;
  onClose: () => void;
  tenantId: string;
  member: NamespaceMember | null;
}) {
  const updateRole = useUpdateMemberRole();
  const [role, setRole] = useState("operator");
  const [submitting, setSubmitting] = useState(false);

  useResetOnOpen(open, () => {
    setRole(member?.role ?? "operator");
    setSubmitting(false);
  });

  const handleSubmit = async () => {
    if (!member) return;
    setSubmitting(true);
    try {
      await updateRole.mutateAsync({ path: { tenant: tenantId, uid: member.id }, body: { role: role as NamespaceMemberRole } });
      onClose();
    } catch {
      /* */
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Role"
      subtitle={
        member ? <span className="font-mono">{member.email}</span> : undefined
      }
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
            disabled={role === member?.role || submitting}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            Save Changes
          </button>
        </>
      )}
    >
      <label className={LABEL}>Role</label>
      <RoleSelector value={role} onChange={setRole} />
    </Drawer>
  );
}

export default EditMemberDrawer;
