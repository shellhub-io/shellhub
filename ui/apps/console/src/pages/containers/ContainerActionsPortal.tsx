import type { UseContainerActionsResult } from "@/hooks/useContainerActions";
import BillingWarning from "@/components/billing/BillingWarning";
import ContainerActionDialog from "./ContainerActionDialog";

interface ContainerActionsPortalProps {
  controller: UseContainerActionsResult;
}

export default function ContainerActionsPortal({ controller }: ContainerActionsPortalProps) {
  const { operation, close } = controller;

  return (
    <>
      <ContainerActionDialog
        key={operation ? `${operation.action}/${operation.entity.uid}` : "closed"}
        open={!!operation}
        container={operation?.entity ?? null}
        action={operation?.action ?? "accept"}
        onClose={close}
        onBillingWarning={controller.onBillingWarning}
        onSuccess={operation ? () => controller.runSuccess(operation.action) : undefined}
      />
      {controller.billingEnabled && (
        <BillingWarning
          open={controller.billingWarningOpen}
          onClose={controller.closeBillingWarning}
        />
      )}
    </>
  );
}
