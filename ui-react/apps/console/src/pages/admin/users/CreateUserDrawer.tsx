import { useState, type FormEvent } from "react";
import { PlusIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useCreateUser } from "@/hooks/useAdminUserMutations";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import Spinner from "@/components/common/Spinner";
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
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={() => void handleSubmit()}
            disabled={!form.isSubmittable || createUser.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {createUser.isPending ? (
              <Spinner tone="onPrimary" />
            ) : (
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
            )}
            Create User
          </button>
        </>
      }
    >
      <form
        onSubmit={(e) => void handleSubmit(e)}
        className="space-y-5"
        noValidate
      >
        <UserFormFields
          form={form}
          idPrefix="create-user"
          autoFocus={open}
        />
        {submitError && (
          <p role="alert" className="text-2xs text-accent-red">
            {submitError}
          </p>
        )}
      </form>
    </Drawer>
  );
}
