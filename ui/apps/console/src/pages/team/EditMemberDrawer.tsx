import { useState } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateMemberRole } from "@/hooks/useMemberMutations";
import { type NamespaceMember } from "@/hooks/useNamespaces";
import Drawer from "@/components/common/Drawer";
import { Button } from "@shellhub/design-system/primitives";
import { RoleSelector } from "./constants";
import { isAssignableRole, type AssignableRole } from "./helpers";

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
  const [role, setRole] = useState<AssignableRole>("operator");
  const [submitting, setSubmitting] = useState(false);

  useResetOnOpen(open, () => {
    setRole(isAssignableRole(member?.role) ? member.role : "operator");
    setSubmitting(false);
  });

  const handleSubmit = async () => {
    if (!member) return;
    setSubmitting(true);
    try {
      await updateRole.mutateAsync({
        path: { tenant: tenantId, uid: member.id },
        body: { role },
      });
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
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={role === member?.role || submitting}
            loading={submitting}
          >
            Save Changes
          </Button>
        </>
      }
    >
      <RoleSelector value={role} onChange={setRole} />
    </Drawer>
  );
}

export default EditMemberDrawer;
