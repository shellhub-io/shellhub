import {
  PencilSquareIcon,
} from "@heroicons/react/24/outline";
import { useAdminEditNamespace } from "@/hooks/useAdminNamespaceMutations";
import { isSdkError } from "@/api/errors";
import FormModal from "@/components/common/FormModal";
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
import ObjectName from "@/components/common/ObjectName";

interface EditNamespaceModalProps {
  open: boolean;
  onClose: () => void;
  namespace: Namespace | null;
}

/**
 * Edits a namespace's limits and settings as an admin, including the ones its own owner cannot
 * change.
 */
export default function EditNamespaceModal({
  open,
  onClose,
  namespace,
}: EditNamespaceModalProps) {
  const editNamespace = useAdminEditNamespace();

  const schema = editNamespaceSchema(namespace?.name ?? "");
  const defaults = buildEditNamespaceDefaults(namespace);

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
    <FormModal
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<PencilSquareIcon />}
      title="Edit namespace"
      description={
        <>
          Rename <ObjectName>{namespace?.name}</ObjectName>, cap its devices or
          turn session recording on and off.
        </>
      }
      submitLabel="Save changes"
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
    </FormModal>
  );
}
