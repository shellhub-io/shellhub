import { useState } from "react";
import { PencilSquareIcon, UserCircleIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { apiErrorMessage } from "@/api/errors";
import { useUpdateMemberRole } from "@/hooks/useMemberMutations";
import ContextualDialog from "@/components/common/ContextualDialog";
import ObjectName from "@/components/common/ObjectName";
import { RoleSelector } from "./constants";
import { assignableRoleOr, type AssignableRole } from "./helpers";

/**
 * Changes a member's role from their row, anchored to the edit button. Saving the role the member
 * already has is not offered, and a role that can't be assigned (an owner's) starts at operator.
 */
export default function MemberRoleDialog({
  tenantId,
  memberId,
  email,
  role,
}: {
  tenantId: string;
  memberId: string;
  email: string;
  role: string;
}) {
  const current = assignableRoleOr(role, "operator");
  const [next, setNext] = useState<AssignableRole>(current);
  const updateRole = useUpdateMemberRole();

  return (
    <ContextualDialog
      width="md"
      trigger={
        <IconButton variant="primary" title="Edit role" aria-label="Edit role">
          <PencilSquareIcon className="w-4 h-4" />
        </IconButton>
      }
      onOpenChange={(open) => {
        if (open) setNext(current);
      }}
      icon={<UserCircleIcon />}
      title="Edit role"
      description={
        <>
          What <ObjectName>{email}</ObjectName> can do in this namespace.
        </>
      }
      submitLabel="Save role"
      submitDisabled={next === current}
      onSubmit={async () => {
        try {
          await updateRole.mutateAsync({
            path: { tenant: tenantId, uid: memberId },
            body: { role: next },
          });
        } catch (err) {
          throw new Error(apiErrorMessage(err));
        }
      }}
    >
      <RoleSelector label="Role" value={next} onChange={setNext} />
    </ContextualDialog>
  );
}
