import { useState, type FormEvent } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateUser } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import { Button } from "@shellhub/design-system/primitives";
import UserFormFields from "./UserFormFields";
import { useUserForm } from "./useUserForm";

interface CreateUserDrawerProps {
  open: boolean;
  onClose: () => void;
}

export default function CreateUserDrawer({
  open,
  onClose,
}: CreateUserDrawerProps) {
  const createUser = useCreateUser();
  const form = useUserForm({ mode: "create" });
  const [submitError, setSubmitError] = useState("");

  useResetOnOpen(open, () => {
    form.reset();
    setSubmitError("");
  });

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!form.validateAll()) return;
    setSubmitError("");
    try {
      await createUser.mutateAsync({ body: form.buildPayload() });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setSubmitError("A user with this email or username already exists.");
      } else {
        setSubmitError("Failed to create user. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Create User"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={!form.isSubmittable || createUser.isPending}
            loading={createUser.isPending}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create User
          </Button>
        </>
      }
    >
      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="space-y-5"
        noValidate
      >
        <UserFormFields form={form} idPrefix="create-user" autoFocus={open} />
        {submitError && (
          <p role="alert" className="text-2xs text-accent-red">
            {submitError}
          </p>
        )}
      </form>
    </Drawer>
  );
}
