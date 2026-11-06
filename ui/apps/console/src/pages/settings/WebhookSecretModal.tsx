import { useQuery } from "@tanstack/react-query";
import { KeyIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import CopyableValue from "@/components/common/CopyableValue";
import Modal from "@/components/common/Modal";
import { webhookRevealSecretOptions } from "@/client";

/**
 * Shows a webhook's signing secret: the one a create just returned, or, when only webhookId is
 * given, the one reveal fetches while the modal is open.
 */
export default function WebhookSecretModal({
  created,
  webhookId,
  onClose,
}: {
  created?: string;
  webhookId?: string | null;
  onClose: () => void;
}) {
  const revealId = created ? null : (webhookId ?? null);
  const revealed = useQuery({
    ...webhookRevealSecretOptions({ path: { id: revealId ?? "" } }),
    enabled: !!revealId,
    gcTime: 0,
  });
  const secret = created ?? revealed.data?.secret ?? "";
  const open = !!created || !!webhookId;

  return (
    <Modal
      open={open}
      onClose={onClose}
      icon={<KeyIcon />}
      title={created ? "Webhook added" : "Webhook secret"}
      description="Your receiver checks each request's X-ShellHub-Signature with this secret."
      footer={
        <Button variant="primary" onClick={onClose}>
          Done
        </Button>
      }
    >
      {revealed.error ? (
        <p role="alert" className="text-2xs text-accent-red">
          Couldn't load the secret.
        </p>
      ) : secret ? (
        <CopyableValue label="Signing secret" value={secret} />
      ) : (
        <p className="text-2xs text-text-muted">Loading the secret…</p>
      )}
    </Modal>
  );
}
