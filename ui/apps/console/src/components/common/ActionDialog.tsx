import { useId, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useHasPermission } from "@/hooks/useHasPermission";
import BaseDialog from "@/components/common/BaseDialog";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { isCloud } from "@/env";
import type { Action, EntityBase, EntityOperation } from "@/hooks/useActionDialog";
import {
  getAcceptErrorMessage,
  isSubscriptionBlocked,
} from "@/utils/acceptErrors";

const VARIANT: Record<EntityOperation, "success" | "warning" | "danger"> = {
  accept: "success",
  reject: "warning",
  remove: "danger",
};

function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

export default function ActionDialog({
  action,
  onClose,
  onSuccess,
  entityType,
  runAction,
}: {
  action: Action;
  onClose: () => void;
  onSuccess: (operation: EntityOperation) => void;
  entityType: "device" | "container";
  runAction: (entity: EntityBase, operation: EntityOperation) => Promise<void>;
}) {
  const navigate = useNavigate();
  const canSubscribe = useHasPermission("billing:subscribe");
  const billingTitleId = useId();
  const [error, setError] = useState<string | null>(null);
  const [billingError, setBillingError] = useState<unknown>(null);

  const { entity, operation } = action;
  const operationLabel = capitalize(operation);
  const entityLabel = capitalize(entityType);

  const handleConfirm = async () => {
    setError(null);
    try {
      await runAction(entity, operation);
    } catch (err: unknown) {
      if (
        operation === "accept" &&
        isSdkError(err) &&
        err.status === 402 &&
        isCloud()
      ) {
        setBillingError(err);
        return;
      }

      if (operation === "accept") setError(getAcceptErrorMessage(err, entityType));
      else setError(`Failed to ${operation} ${entityType}.`);

      return;
    }
    onSuccess(operation);
    onClose();
  };

  if (billingError) {
    const billingTitle = isSubscriptionBlocked(billingError)
      ? "Subscription issue"
      : `${entityLabel} limit reached`;
    const billingMessage = getAcceptErrorMessage(billingError, entityType, canSubscribe);

    if (canSubscribe) {
      return (
        <ConfirmDialog
          open
          onClose={onClose}
          onConfirm={() => {
            void navigate("/settings#billing");
            onClose();
          }}
          title={billingTitle}
          description={billingMessage}
          variant="warning"
          confirmLabel="Go to billing"
          cancelLabel="Not now"
        />
      );
    }

    return (
      <BaseDialog open onClose={onClose} aria-labelledby={billingTitleId}>
        <div className="p-6 pb-0">
          <h2 id={billingTitleId} className="text-base font-semibold text-text-primary">
            {billingTitle}
          </h2>
        </div>
        <div className="px-6 pt-2 pb-6">
          <p className="text-sm text-text-muted">{billingMessage}</p>
        </div>
        <div className="flex justify-end px-6 py-4 border-t border-border">
          <Button variant="ghost" onClick={onClose}>Close</Button>
        </div>
      </BaseDialog>
    );
  }

  const description = (
    <>
      Do you want to {operation}{" "}
      <span className="font-medium text-text-primary">{entity.name}</span>?
      {operation === "remove" && (
        <p className="text-xs text-text-muted/70 mt-1">
          This action cannot be undone.
        </p>
      )}
    </>
  );

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={handleConfirm}
      title={`${operationLabel} ${entityLabel}`}
      description={description}
      confirmLabel={operationLabel}
      variant={VARIANT[operation]}
      errorMessage={error}
    />
  );
}
