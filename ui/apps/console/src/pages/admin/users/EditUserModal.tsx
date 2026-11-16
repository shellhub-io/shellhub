import { useMutation } from "@tanstack/react-query";
import { PencilSquareIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useHandOverSignOut } from "@/hooks/useHandOverSignOut";
import { apiErrorFields, isSdkError } from "@/api/errors";
import { emailRefusalMessage } from "@/utils/emailRefusal";
import FormModal from "@/components/common/FormModal";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import UserFormFields from "./UserFormFields";
import {
  userSchema,
  buildUserDefaults,
  buildUserPayload,
  type UserFormValues,
} from "./userSchema";
import { adminUpdateUserMutation, type UserAdminResponse } from "@/client";
import ObjectName from "@/components/common/ObjectName";

const ADMIN_EMAIL_REFUSAL_MESSAGES: Record<string, string> = {
  "must be different from the recovery email":
    "Must be different from the user's recovery email",
};

interface EditUserModalProps {
  open: boolean;
  onClose: () => void;
  user: UserAdminResponse | null;
}

/**
 * Edits a user as an admin, including the namespace limits the user cannot change themselves, and
 * calls onClose once saved. An admin who sets a new password on their own account revokes the
 * session's token, so the modal then hands the sign-out over to the sign-in page and leaves for it
 * instead of calling onClose. When saving fails, own password included, the modal stays open with
 * the error and the session is left as it was.
 */
export default function EditUserModal({
  open,
  onClose,
  user,
}: EditUserModalProps) {
  const invalidateUser = useInvalidateByIds("getUsers", "getUser");
  const updateUser = useMutation({
    ...adminUpdateUserMutation(),
    onSuccess: invalidateUser,
  });
  const currentUserId = useAuthStore((s) => s.userId);
  const handOverSignOut = useHandOverSignOut();

  const usernameRequired = !!user?.username;
  const schema = userSchema("edit", usernameRequired);
  const defaults = buildUserDefaults(user);

  const form = useDrawerForm(open, schema, defaults);
  const { control, setError, clearErrors } = form;

  const isSelf = user?.id === currentUserId;
  const canChangeConfirmed = user?.status !== "confirmed";
  const disableAdmin = !!user?.admin && isSelf;

  const onValid = async (values: UserFormValues) => {
    if (!user) return;
    clearErrors("root");
    const update = () =>
      updateUser.mutateAsync({
        path: { id: user.id },
        body: buildUserPayload("edit", values, user),
      });
    try {
      if (isSelf && values.password !== "") {
        await handOverSignOut(
          update,
          "Password changed. Sign in with your new password.",
        );
      } else {
        await update();
        onClose();
      }
    } catch (err) {
      const emailReason = apiErrorFields(err).email;
      if (emailReason) {
        setError("email", {
          type: "server",
          message:
            emailRefusalMessage(emailReason) ??
            ADMIN_EMAIL_REFUSAL_MESSAGES[emailReason] ??
            "This email cannot be used.",
        });
        return;
      }

      const message =
        isSdkError(err) && err.status === 409
          ? "A user with this email or username already exists."
          : "Failed to update user. Please try again.";

      setError("root", { message });
    }
  };

  return (
    <FormModal
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<PencilSquareIcon />}
      title="Edit user"
      description={
        <>
          Change <ObjectName>{user?.username || user?.email}</ObjectName>'s
          details, password and whether they administer the instance.
        </>
      }
      submitLabel="Save changes"
    >
      <UserFormFields
        control={control}
        mode="edit"
        idPrefix="edit-user"
        canChangeConfirmed={canChangeConfirmed}
        disableAdmin={disableAdmin}
        usernameRequired={usernameRequired}
      />
    </FormModal>
  );
}
