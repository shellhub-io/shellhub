import { useId, useState } from "react";
import { Button, Card } from "@shellhub/design-system/primitives";
import { type InstallKeyEvent } from "@/client";
import CopyButton from "@/components/common/CopyButton";
import BaseDialog from "@/components/common/BaseDialog";
import { LABEL } from "@/utils/styles";
import KeyValueChip from "./KeyValueChip";

/**
 * The registered device's key fingerprint, shown inline in the row; clicking it opens the full device
 * key (PEM) the device presented when it registered. "Device key" (not "public key", which reads as an
 * SSH user key) is the device's own identity key that authenticates its agent to the server. It is not
 * a secret, so it is shown directly with a copy control, no reveal gate. Renders nothing with no key.
 */
export default function EventPublicKey({ event }: { event: InstallKeyEvent }) {
  const [open, setOpen] = useState(false);
  const autoId = useId();
  const titleId = `event-pubkey-title-${autoId}`;
  const keyLabelId = `event-pubkey-key-${autoId}`;

  if (!event.fingerprint) return null;

  return (
    <>
      <KeyValueChip
        label="Key"
        value={
          <span className="block max-w-[170px] truncate">
            {event.fingerprint}
          </span>
        }
        onClick={() => setOpen(true)}
        title="View device key"
        ariaLabel="View device key"
      />

      <BaseDialog
        open={open}
        onClose={() => setOpen(false)}
        size="md"
        aria-labelledby={titleId}
      >
        <div className="p-6 pb-0">
          <h2
            id={titleId}
            className="text-base font-semibold text-text-primary"
          >
            {event.hostname || "Device key"}
          </h2>
          <p className="mt-1 text-sm text-text-muted">
            The device's own identity key, presented when it registered — not a
            secret. Safe to share.
          </p>
        </div>

        <div className="space-y-4 px-6 pb-6 pt-4">
          <div>
            <span className={LABEL}>Fingerprint</span>
            <Card className="flex items-center gap-2 rounded-lg px-3.5 py-2.5">
              <code
                className="flex-1 truncate text-xs font-mono text-text-muted select-all"
                title={event.fingerprint}
              >
                {event.fingerprint}
              </code>
              <CopyButton text={event.fingerprint} size="md" />
            </Card>
          </div>

          <div>
            <span id={keyLabelId} className={LABEL}>
              Device key
            </span>
            <Card
              aria-labelledby={keyLabelId}
              className="flex items-start gap-2 rounded-lg px-3.5 py-2.5"
            >
              <code className="max-h-48 flex-1 overflow-y-auto whitespace-pre-wrap break-all text-xs font-mono text-text-muted select-all">
                {event.public_key}
              </code>
              <CopyButton text={event.public_key ?? ""} size="md" />
            </Card>
          </div>
        </div>

        <div className="flex justify-end gap-2 border-t border-border px-6 py-4">
          <Button variant="ghost" onClick={() => setOpen(false)}>
            Close
          </Button>
        </div>
      </BaseDialog>
    </>
  );
}
