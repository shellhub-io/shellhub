import type { UseDeviceActionsResult } from "@/hooks/useDeviceActions";
import BillingWarning from "@/components/billing/BillingWarning";
import DeviceActionDialog from "./DeviceActionDialog";

interface DeviceActionsPortalProps {
  controller: UseDeviceActionsResult;
}

export default function DeviceActionsPortal({ controller }: DeviceActionsPortalProps) {
  const { operation, close } = controller;

  return (
    <>
      <DeviceActionDialog
        key={operation ? `${operation.action}/${operation.entity.uid}` : "closed"}
        open={!!operation}
        device={operation?.entity ?? null}
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
