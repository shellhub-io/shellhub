import { useState } from "react";
import { Link } from "react-router-dom";
import {
  ArrowRightStartOnRectangleIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import {
  type Namespace,
  useDeleteNamespace,
  useLeaveNamespace,
} from "@/hooks/useNamespaceRequests";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useMemberDevices } from "@/hooks/useMemberDevices";
import { useAuthStore } from "@/stores/authStore";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DepartingDevices from "@/components/common/DepartingDevices";
import InputField from "@/components/common/fields/InputField";
import SettingsDangerCard from "@/components/settings/SettingsDangerCard";

type DeleteFailure = "bound" | "billing" | "other";

function deleteFailure(err: unknown): DeleteFailure {
  if (!isSdkError(err)) return "other";
  if (err.status === 409) return "bound";
  if (err.status === 402) return "billing";
  return "other";
}

function DeleteFailureMessage({ failure }: { failure: DeleteFailure }) {
  if (failure === "bound") {
    return "This namespace is bound to the instance and can't be deleted.";
  }
  if (failure === "billing") {
    return (
      <>
        Billing is blocking this deletion: the namespace has an active
        subscription or an unpaid invoice. Resolve it in{" "}
        <Link to="/settings/billing" className="underline hover:no-underline">
          billing settings
        </Link>
        , then try again.
      </>
    );
  }
  return "Couldn't delete the namespace.";
}

function DeleteDialog({
  namespaceName,
  tenantId,
  onClose,
}: {
  namespaceName: string;
  tenantId: string;
  onClose: () => void;
}) {
  const deleteNs = useDeleteNamespace();
  const [confirm, setConfirm] = useState("");
  const [failure, setFailure] = useState<DeleteFailure | null>(null);

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setFailure(null);
        try {
          await deleteNs.mutateAsync(tenantId);
        } catch (err) {
          setFailure(deleteFailure(err));
          throw new Error();
        }
      }}
      icon={<TrashIcon />}
      title="Delete namespace"
      description="Its devices, sessions, keys and configuration are removed. This can't be undone."
      confirmLabel="Delete namespace"
      confirmDisabled={confirm !== namespaceName}
    >
      <div className="mb-4">
        <InputField
          id="delete-ns-confirm"
          label={`Type "${namespaceName}" to confirm`}
          value={confirm}
          onChange={setConfirm}
          placeholder={namespaceName}
        />
      </div>
      {failure && (
        <p className="text-2xs text-accent-red mb-3">
          <DeleteFailureMessage failure={failure} />
        </p>
      )}
    </ConfirmDialog>
  );
}

/**
 * Confirms leaving the namespace, then leaves it. A failed leave shows its error and keeps the
 * dialog open, since the member is still in the namespace and may try again.
 */
export function LeaveDialog({
  tenantId,
  onClose,
}: {
  tenantId: string;
  onClose: () => void;
}) {
  const leaveNs = useLeaveNamespace();
  const userId = useAuthStore((s) => s.userId);
  const devices = useMemberDevices(userId ?? "");
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setError("");
        try {
          await leaveNs.mutateAsync(tenantId);
        } catch {
          setError("Couldn't leave the namespace.");
          throw new Error();
        }
      }}
      icon={<ArrowRightStartOnRectangleIcon />}
      title="Leave namespace"
      description="You lose access to its devices and sessions. To rejoin, someone has to invite you again."
      confirmLabel="Leave namespace"
      confirmDisabled={devices.isLoading || devices.isError}
    >
      {userId && <DepartingDevices memberId={userId} self />}
      {error && <p className="text-2xs text-accent-red mb-3">{error}</p>}
    </ConfirmDialog>
  );
}

/**
 * Deleting the namespace, for whoever may: a red card that asks for the name to be typed before it
 * goes. Renders nothing for anyone without the permission, who leaves instead.
 */
export function DeleteNamespaceZone({ ns }: { ns: Namespace }) {
  const canDelete = useHasPermission("namespace:delete");
  const [open, setOpen] = useState(false);

  if (!canDelete) return null;

  return (
    <div className="pt-4">
      <p className="mb-3 text-2xs font-mono uppercase tracking-label text-accent-red">
        Danger zone
      </p>
      <SettingsDangerCard
        title="Delete namespace"
        description="Removes its devices, sessions, keys and configuration for good."
        action={
          <Button size="sm" variant="destructive" onClick={() => setOpen(true)}>
            Delete namespace
          </Button>
        }
      />

      {open && (
        <DeleteDialog
          namespaceName={ns.name}
          tenantId={ns.tenant_id}
          onClose={() => setOpen(false)}
        />
      )}
    </div>
  );
}
