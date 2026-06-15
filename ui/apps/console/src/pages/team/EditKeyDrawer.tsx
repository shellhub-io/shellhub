import { useState } from "react";
import { ExclamationCircleIcon } from "@heroicons/react/24/outline";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useUpdateApiKey } from "@/hooks/useApiKeyMutations";
import { type ApiKey } from "@/client";
import Drawer from "@/components/common/Drawer";
import InputField from "@/components/common/fields/InputField";
import { Button } from "@shellhub/design-system/primitives";
import { RoleSelector } from "./constants";
import { isAssignableRole, type AssignableRole } from "./helpers";

/* ─── Edit API Key Drawer ─── */

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
  const [name, setName] = useState("");
  const [role, setRole] = useState<AssignableRole>("administrator");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    setName(apiKey?.name ?? "");
    setRole(isAssignableRole(apiKey?.role) ? apiKey.role : "administrator");
    setSubmitting(false);
    setError(null);
  });

  const handleSubmit = async () => {
    if (!apiKey || !name.trim()) return;
    setSubmitting(true);
    setError(null);
    try {
      await updateKey.mutateAsync({
        path: { key: apiKey.name },
        body: { name: name.trim(), role },
      });
      onClose();
    } catch (err: unknown) {
      setError(
        err instanceof Error ? err.message : "Failed to update API key.",
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Edit API Key"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit()}
            disabled={!name.trim() || submitting}
            loading={submitting}
          >
            Save Changes
          </Button>
        </>
      }
    >
      <div className="space-y-5">
        <InputField
          id="edit-key-name"
          label="Name"
          value={name}
          onChange={setName}
          maxLength={20}
          autoFocus={open}
        />
        <RoleSelector value={role} onChange={setRole} />

        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}
      </div>
    </Drawer>
  );
}

export default EditKeyDrawer;
