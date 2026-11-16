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
import { useReplacedDevice } from "@/hooks/useReplacedDevice";
import BaseDialog from "@/components/common/BaseDialog";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import ObjectName from "@/components/common/ObjectName";
import { features } from "@/edition";
import type {
  Action,
  EntityBase,
  EntityOperation,
} from "@/hooks/useActionDialog";
import { getAcceptErrorMessage } from "@/utils/acceptErrors";
import { useCurrentNamespace } from "@/hooks/useNamespaceRequests";
import { isSubscriptionBlocked } from "@/utils/billing";
import { capitalize } from "@/utils/string";
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
      <>
        {name} is removed from the namespace. While its agent keeps running, it
        can enroll again.
      </>
    ),
  },
};

/**
 * The confirmation for an accept, reject or remove. It takes the action rather than a flag, so
 * the wording, the icon, the button and the danger follow from one value and cannot disagree,
 * except an accept that would replace an offline device or container with the same identity: that
 * one confirms the replacement as a danger, disabled without DeviceRemove. Every accept is held
 * while that lookup runs, refused once it fails, and refused while the device with the same
 * identity is online.
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
  runAction: (
    entity: EntityBase,
    operation: EntityOperation,
    replace: boolean,
  ) => Promise<void>;
}) {
  const navigate = useNavigate();
  const { namespace } = useCurrentNamespace();
  const hasSubscription = isSubscriptionBlocked(namespace?.billing);
  const canSubscribe = useHasPermission("BillingCreateSubscription");
  const billingTitleId = useId();
  const [error, setError] = useState<string | null>(null);
  const [billingError, setBillingError] = useState<unknown>(null);

  const { entity, operation } = action;
  const { replaced, acceptBlocked, blockedMessage } = useReplacedDevice(
    operation === "accept" ? entity.uid : "",
  );
  const canRemove = useHasPermission("DeviceRemove");
  const operationLabel = capitalize(operation);
  const entityLabel = capitalize(entityType);

  const handleConfirm = async () => {
    setError(null);
    try {
      await runAction(entity, operation, !!replaced);
    } catch (err: unknown) {
      if (
        operation === "accept" &&
        isSdkError(err) &&
        err.status === 402 &&
        features().billing
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

  if (replaced) {
    const replacedName = <ObjectName>{replaced.name}</ObjectName>;

    return (
      <ConfirmDialog
        open
        onClose={onClose}
        onConfirm={handleConfirm}
        icon={<ExclamationTriangleIcon />}
        title={`${operationLabel} ${entityType}`}
        description={
          canRemove ? (
            <>
              {replacedName} is offline and reports the same identity. Accepting
              replaces it. <ObjectName>{entity.name}</ObjectName> takes its name
              and session history, and ShellHub deletes the record of{" "}
              {replacedName}.
            </>
          ) : (
            <>
              {replacedName} is offline and reports the same identity. Accepting{" "}
              <ObjectName>{entity.name}</ObjectName> would replace it and delete
              its record, which only a member allowed to remove devices can do.
            </>
          )
        }
        confirmLabel={`Replace ${replaced.name}`}
        confirmDisabled={!canRemove}
        variant="danger"
        errorMessage={error}
      />
    );
  }

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={handleConfirm}
      icon={icon}
      title={`${operationLabel} ${entityType}`}
      description={outcome(<ObjectName>{entity.name}</ObjectName>)}
      confirmLabel={`${operationLabel} ${entityType}`}
      confirmDisabled={acceptBlocked}
      variant={variant}
      errorMessage={error ?? blockedMessage}
    />
  );
}
