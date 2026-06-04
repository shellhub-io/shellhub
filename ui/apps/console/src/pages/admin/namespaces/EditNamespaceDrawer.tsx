import { useState, type FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useAdminEditNamespace } from "@/hooks/useAdminNamespaceMutations";
import { isSdkError } from "@/api/errors";
import Drawer from "@/components/common/Drawer";
import NamespaceNameField from "@/components/common/fields/NamespaceNameField";
import NumericInput from "@/components/common/fields/NumericInput";
import CheckboxField from "@/components/common/fields/CheckboxField";
import { validateNamespaceName } from "@/utils/validation";
import type { Namespace } from "@/client";
import Spinner from "@/components/common/Spinner";

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

  const [name, setName] = useState("");
  const [maxDevices, setMaxDevices] = useState(() =>
    String(namespace?.max_devices ?? -1),
  );
  const [sessionRecord, setSessionRecord] = useState(false);
  const [deviceAutoAccept, setDeviceAutoAccept] = useState(false);
  const [error, setError] = useState("");

  useResetOnOpen(open, () => {
    setName(namespace?.name ?? "");
    setMaxDevices(String(namespace?.max_devices ?? -1));
    setSessionRecord(namespace?.settings?.session_record ?? false);
    setDeviceAutoAccept(namespace?.settings?.device_auto_accept ?? false);
    setError("");
  });

  const isMaxDevicesValid = parseInt(maxDevices, 10) >= -1;
  const nameValidationError =
    name !== namespace?.name ? validateNamespaceName(name) : null;
  const canSubmit = !nameValidationError && isMaxDevicesValid;

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (!canSubmit || !namespace) return;
    setError("");
    try {
      await editNamespace.mutateAsync({
        path: { tenantID: namespace.tenant_id },
        // The SDK types body as full Namespace; we spread the original
        // to satisfy the type while only changing the editable fields.
        body: {
          ...namespace,
          name: name.trim(),
          max_devices: parseInt(maxDevices, 10),
          settings: {
            connection_announcement:
              namespace.settings?.connection_announcement ?? "",
            session_record: sessionRecord,
            device_auto_accept: deviceAutoAccept,
          },
        },
      });
      onClose();
    } catch (err) {
      if (isSdkError(err) && err.status === 409) {
        setError("A namespace with this name already exists.");
      } else {
        setError("Failed to update namespace. Please try again.");
      }
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit Namespace"
      subtitle={
        namespace ? (
          <span className="font-mono">{namespace.name}</span>
        ) : undefined
      }
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
            type="button"
            onClick={() => void handleSubmit()}
            disabled={!canSubmit || editNamespace.isPending}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all flex items-center gap-2"
          >
            {editNamespace.isPending && (
              <Spinner tone="onPrimary" />
            )}
            Save Changes
          </button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        <NamespaceNameField
          id="edit-ns-name"
          value={name}
          onChange={setName}
          autoFocus={open}
          error={nameValidationError}
        />

        <NumericInput
          id="edit-ns-max-devices"
          label="Max Devices"
          value={maxDevices}
          onChange={setMaxDevices}
          allowNegative
          hint="Use -1 for unlimited devices"
          error={
            isMaxDevicesValid
              ? undefined
              : "Max devices must be a number greater than or equal to -1"
          }
        />

        <CheckboxField
          id="edit-namespace-session-record"
          label="Session Recording"
          checked={sessionRecord}
          onChange={setSessionRecord}
        />

        <CheckboxField
          id="edit-namespace-device-auto-accept"
          label="Auto-Accept Devices"
          checked={deviceAutoAccept}
          onChange={setDeviceAutoAccept}
        />

        {error && (
          <p role="alert" className="text-2xs text-accent-red">
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}
