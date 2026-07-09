import { useMemo } from "react";
import { useUpdateApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKey } from "@/client";
import FormDrawer from "@/components/common/FormDrawer";
import { FormInputField } from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { FormRoleSelector } from "./constants";
import {
  editKeySchema,
  buildEditKeyDefaults,
  buildEditKeyBody,
  type EditKeyFormValues,
} from "./schemas";

function EditKeyDrawer({
  open,
  onClose,
  apiKey,
}: {
  open: boolean;
  onClose: () => void;
  apiKey: ApiKey | null;
}) {
  const updateKey = useUpdateApiKey();
  const defaults = useMemo(() => buildEditKeyDefaults(apiKey), [apiKey]);
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
    <FormDrawer
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      title="Edit API Key"
      submitLabel="Save Changes"
    >
      <FormInputField
        name="name"
        control={control}
        id="edit-key-name"
        label="Name"
        maxLength={20}
      />
      <FormRoleSelector name="role" control={control} />
    </FormDrawer>
  );
}

export default EditKeyDrawer;
