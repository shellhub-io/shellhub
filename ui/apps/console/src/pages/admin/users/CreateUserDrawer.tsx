import { useMemo } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useCreateUser } from "@/hooks/useAdminUserMutations";
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

interface CreateUserDrawerProps {
  open: boolean;
  onClose: () => void;
}

export default function CreateUserDrawer({
  open,
  onClose,
}: CreateUserDrawerProps) {
  const createUser = useCreateUser();

  const schema = useMemo(() => userSchema("create"), []);
  const defaults = useMemo(() => buildUserDefaults(), []);

  const form = useDrawerForm(open, schema, defaults);
  const { control, setError, clearErrors } = form;

  const onValid = async (values: UserFormValues) => {
    clearErrors("root");
    try {
      await createUser.mutateAsync({ body: buildUserPayload("create", values) });
      onClose();
    } catch (err) {
      const message = isSdkError(err) && err.status === 409
        ? "A user with this email or username already exists."
        : "Failed to create user. Please try again.";

      setError("root", { message });
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      title="Create User"
      submitLabel="Create User"
      submitIcon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <UserFormFields control={control} mode="create" idPrefix="create-user" />
    </FormDrawer>
  );
}
