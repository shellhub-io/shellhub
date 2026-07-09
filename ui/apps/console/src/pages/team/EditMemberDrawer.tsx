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

  const onValid = async (values: EditRoleFormValues) => {
    if (!member) return;
    try {
      await updateRole.mutateAsync({
        path: { tenant: tenantId, uid: member.id },
        body: { role: values.role },
      });
      onClose();
    } catch {
      /* Role changes are low-stakes; surface nothing and let the user retry. */
    }
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
