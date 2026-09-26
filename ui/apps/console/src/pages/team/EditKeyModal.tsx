import {
  PencilSquareIcon,
} from "@heroicons/react/24/outline";
import { useUpdateApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKey } from "@/client";
import FormModal from "@/components/common/FormModal";
import { FormInputField } from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { FormRoleSelector } from "./constants";
import {
  editKeySchema,
  buildEditKeyDefaults,
  buildEditKeyBody,
  type EditKeyFormValues,
} from "./schemas";

/**
 * Edits an API key's name and role. The secret is not re-issued, so an edit cannot recover a key
 * that was never saved.
 */
function EditKeyModal({
  open,
  onClose,
  apiKey,
}: {
  open: boolean;
  onClose: () => void;
  apiKey: ApiKey | null;
}) {
  const updateKey = useUpdateApiKey();
  const defaults = buildEditKeyDefaults(apiKey);
  const form = useDrawerForm(open, editKeySchema, defaults);
  const { control, setError, clearErrors } = form;

  const onValid = async (values: EditKeyFormValues) => {
    if (!apiKey) return;
    clearErrors("root");
    try {
      await updateKey.mutateAsync({
        path: { key: apiKey.name },
        body: buildEditKeyBody(values),
      });
      onClose();
    } catch (err: unknown) {
      setError("root", {
        message:
          err instanceof Error ? err.message : "Failed to update API key.",
      });
    }
  };

  return (
    <FormModal
      size="sm"
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<PencilSquareIcon />}
      title="Edit API key"
      description="Change the key's name and role. Its value stays the same, so integrations using it keep working."
      submitLabel="Save changes"
    >
      <FormInputField
        name="name"
        control={control}
        id="edit-key-name"
        label="Name"
        maxLength={20}
      />
      <FormRoleSelector name="role" control={control} />
    </FormModal>
  );
}

export default EditKeyModal;
