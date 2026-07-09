import { useMemo } from "react";
import { useUpdateUser } from "@/hooks/useAdminUserMutations";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import FormDrawer from "@/components/common/FormDrawer";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import UserFormFields from "./UserFormFields";
import {
  userSchema,
  buildUserDefaults,
  buildUserPayload,
  type UserFormValues,
} from "./userSchema";
import type { UserAdminResponse } from "@/client";

interface EditUserDrawerProps {
  open: boolean;
  onClose: () => void;
  user: UserAdminResponse | null;
}

export default function EditUserDrawer({
  open,
  onClose,
  user,
}: EditUserDrawerProps) {
  const updateUser = useUpdateUser();
  const currentUsername = useAuthStore((s) => s.username);

  const schema = useMemo(() => userSchema("edit"), []);
  const defaults = useMemo(() => buildUserDefaults(user), [user]);

  const form = useDrawerForm(open, schema, defaults);
  const { control, setError, clearErrors } = form;

  const isSelf = user?.username === currentUsername;
  const canChangeConfirmed = user?.status !== "confirmed";
  const disableAdmin = !!user?.admin && isSelf;

  const onValid = async (values: UserFormValues) => {
    if (!user) return;
    clearErrors("root");
    try {
      await updateUser.mutateAsync({
        path: { id: user.id },
        body: buildUserPayload("edit", values, user),
      });
      onClose();
    } catch (err) {
      const message = isSdkError(err) && err.status === 409
        ? "A user with this email or username already exists."
        : "Failed to update user. Please try again.";

      setError("root", { message });
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      title="Edit User"
      submitLabel="Save Changes"
      subtitle={
        user ? <span className="font-mono">{user.username}</span> : undefined
      }
    >
      <UserFormFields
        control={control}
        mode="edit"
        idPrefix="edit-user"
        canChangeConfirmed={canChangeConfirmed}
        disableAdmin={disableAdmin}
      />
    </FormDrawer>
  );
}
