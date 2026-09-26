import { useId, useState } from "react";
import type { ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import {
  CheckCircleIcon,
  ExclamationTriangleIcon,
  NoSymbolIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isSdkError } from "@/api/errors";
import { useHasPermission } from "@/hooks/useHasPermission";
import BaseDialog from "@/components/common/BaseDialog";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";
import { isCloud } from "@/env";
import type {
  Action,
  EntityBase,
  EntityOperation,
} from "@/hooks/useActionDialog";
import { getAcceptErrorMessage } from "@/utils/acceptErrors";
import { useNamespace } from "@/hooks/useNamespaces";
import { useAuthStore } from "@/stores/authStore";
import { isSubscriptionBlocked } from "@/utils/billing";
import DialogHeader from "@/components/common/DialogHeader";

const OPERATION: Record<
  EntityOperation,
  {
    variant: "success" | "warning" | "danger";
    icon: ReactNode;
    outcome: (name: ReactNode) => ReactNode;
  }
> = {
  accept: {
    variant: "success",
    icon: <CheckCircleIcon />,
    outcome: (name) => (
      <>{name} joins the namespace and can be reached over SSH.</>
    ),
  },
  reject: {
    variant: "warning",
    icon: <NoSymbolIcon />,
    outcome: (name) => (
      <>{name} stays out of the namespace and can't be reached.</>
    ),
  },
  remove: {
    variant: "danger",
    icon: <TrashIcon />,
    outcome: (name) => (
      <>{name} is removed from the namespace. This can't be undone.</>
    ),
  },
};

function capitalize(str: string) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}

/**
 * The confirmation for an accept, reject or remove. It takes the action rather than a flag, so
 * the wording, the icon, the button and the danger all follow from one value and cannot disagree.
 */
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
  const tenant = useAuthStore((s) => s.tenant);
  const { namespace } = useNamespace(tenant ?? "");
  const hasSubscription = isSubscriptionBlocked(namespace?.billing);
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

      if (operation === "accept")
        setError(
          getAcceptErrorMessage(err, hasSubscription, canSubscribe, entityType),
        );
      else setError(`Failed to ${operation} ${entityType}.`);

      return;
    }
    onSuccess(operation);
    onClose();
  };

  if (billingError) {
    const billingTitle = hasSubscription
      ? "Subscription issue"
      : `${entityLabel} limit reached`;
    const billingMessage = getAcceptErrorMessage(
      billingError,
      hasSubscription,
      canSubscribe,
      entityType,
    );

    if (canSubscribe) {
      return (
        <ConfirmDialog
          open
          onClose={onClose}
          onConfirm={() => {
            void navigate("/settings/billing");
            onClose();
          }}
          icon={<ExclamationTriangleIcon />}
          title={billingTitle}
          description={billingMessage}
          variant="warning"
          confirmLabel="Go to billing"
          cancelLabel="Not now"
        />
      );
    }

    return (
      <BaseDialog
        open
        onClose={onClose}
        aria-labelledby={billingTitleId}
        aria-describedby={`${billingTitleId}-description`}
      >
        <DialogHeader
          icon={<ExclamationTriangleIcon />}
          iconColor="yellow"
          title={billingTitle}
          description={billingMessage}
          titleId={billingTitleId}
          descriptionId={`${billingTitleId}-description`}
        />
        <div className="flex justify-end px-6 py-4 border-t border-border">
          <Button variant="ghost" onClick={onClose}>
            Close
          </Button>
        </div>
      </BaseDialog>
    );
  }

  const { variant, icon, outcome } = OPERATION[operation];

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={handleConfirm}
      icon={icon}
      title={`${operationLabel} ${entityType}`}
      description={outcome(<ObjectName>{entity.name}</ObjectName>)}
      confirmLabel={`${operationLabel} ${entityType}`}
      variant={variant}
      errorMessage={error}
    />
  );
}
