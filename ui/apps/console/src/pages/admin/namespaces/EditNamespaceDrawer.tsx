import { useMemo } from "react";
import { useAdminEditNamespace } from "@/hooks/useAdminNamespaceMutations";
import { isSdkError } from "@/api/errors";
import FormDrawer from "@/components/common/FormDrawer";
import {
  FormInputField,
  FormNumericInput,
  FormCheckboxField,
} from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import {
  NAMESPACE_NAME_HINT,
  NAMESPACE_NAME_MAX_LENGTH,
} from "@/utils/validation";
import {
  editNamespaceSchema,
  buildEditNamespaceDefaults,
  buildEditNamespaceBody,
  type EditNamespaceFormValues,
} from "./editNamespaceSchema";
import type { Namespace } from "@/client";

interface EditNamespaceDrawerProps {
  open: boolean;
  onClose: () => void;
  namespace: Namespace | null;
}

export default function EditNamespaceDrawer({
  open,
  onClose,
  namespace,
}: EditNamespaceDrawerProps) {
  const editNamespace = useAdminEditNamespace();

  const schema = useMemo(
    () => editNamespaceSchema(namespace?.name ?? ""),
    [namespace?.name],
  );
  const defaults = useMemo(
    () => buildEditNamespaceDefaults(namespace),
    [namespace],
  );

  const form = useDrawerForm(open, schema, defaults);
  const { control, setValue, setError, clearErrors } = form;

  const onValid = async (values: EditNamespaceFormValues) => {
    if (!namespace) return;
    clearErrors("root");
    try {
      await editNamespace.mutateAsync({
        path: { tenantID: namespace.tenant_id },
        body: buildEditNamespaceBody(namespace, values),
      });
      onClose();
    } catch (err) {
      const message =
        isSdkError(err) && err.status === 409
          ? "A namespace with this name already exists."
          : "Failed to update namespace. Please try again.";

      setError("root", { message });
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      title="Edit Namespace"
      submitLabel="Save Changes"
      subtitle={
        namespace ? (
          <span className="font-mono">{namespace.name}</span>
        ) : undefined
      }
    >
      <FormInputField
        name="name"
        control={control}
        id="edit-ns-name"
        label="Namespace Name"
        placeholder="my-namespace"
        hint={NAMESPACE_NAME_HINT}
        maxLength={NAMESPACE_NAME_MAX_LENGTH}
        onValueChange={(v) =>
          setValue("name", v.toLowerCase(), {
            shouldDirty: true,
            shouldValidate: true,
          })
        }
      />

      <FormNumericInput
        name="maxDevices"
        control={control}
        id="edit-ns-max-devices"
        label="Max Devices"
        allowNegative
        hint="Use -1 for unlimited devices"
      />

      <FormCheckboxField
        name="sessionRecord"
        control={control}
        id="edit-namespace-session-record"
        label="Session Recording"
      />

    </FormDrawer>
  );
}
