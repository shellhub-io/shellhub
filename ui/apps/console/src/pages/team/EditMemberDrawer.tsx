import { useMemo } from "react";
import { useUpdateMemberRole } from "@/hooks/useMemberMutations";
import { type NamespaceMember } from "@/hooks/useNamespaces";
import FormDrawer from "@/components/common/FormDrawer";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { FormRoleSelector } from "./constants";
import {
  editRoleSchema,
  buildMemberRoleDefaults,
  type EditRoleFormValues,
} from "./schemas";

/**
 * Changes a member's role. Failures are not surfaced here — a role change is low-stakes and the
 * list simply refetches — which is why the drawer closes either way.
 */
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
  const defaults = useMemo(() => buildMemberRoleDefaults(member), [member]);
  const form = useDrawerForm(open, editRoleSchema, defaults);

  const onValid = (values: EditRoleFormValues) => {
    if (!member) return;
    updateRole.mutate(
      {
        path: { tenant: tenantId, uid: member.id },
        body: { role: values.role },
      },
      { onSuccess: onClose },
    );
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      title="Edit Role"
      submitLabel="Save Changes"
      requireDirty
      subtitle={
        member ? <span className="font-mono">{member.email}</span> : undefined
      }
    >
      <FormRoleSelector name="role" control={form.control} />
    </FormDrawer>
  );
}

export default EditMemberDrawer;
