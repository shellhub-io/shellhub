import { useState } from "react";
import BaseDialog from "./BaseDialog";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";

/**
 * Modal that runs the whole pairing flow in place — enter code, resolve,
 * preview, pick a namespace, accept — without ever leaving the Add Device page.
 * The code is kept here and cleared on close so reopening starts fresh.
 */
export default function PairingCodeDialog({
  open,
  onClose,
}: {
  open: boolean;
  onClose: () => void;
}) {
  const [code, setCode] = useState("");

  const handleClose = () => {
    setCode("");
    onClose();
  };

  return (
    <BaseDialog
      open={open}
      onClose={handleClose}
      size="md"
      aria-label="Claim a device"
    >
      <div className="p-6">
        <AcceptDeviceFlow code={code} onCodeChange={setCode} />
      </div>
    </BaseDialog>
  );
}
