import { useQuery } from "@tanstack/react-query";
import { KeyIcon } from "@heroicons/react/24/outline";
import { Button, Card } from "@shellhub/design-system/primitives";
import CopyButton from "@/components/common/CopyButton";
import Modal from "@/components/common/Modal";
import { webhookRevealSecretOptions } from "@/client";
import { LABEL } from "@/utils/styles";

function SecretCard({ secret }: { secret: string }) {
  return (
    <div>
      <span id="webhook-secret-label" className={LABEL}>
        Signing secret
      </span>
      <Card
        aria-labelledby="webhook-secret-label"
        className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
      >
        <code className="flex-1 text-xs font-mono text-accent-cyan break-all select-all">
          {secret}
        </code>
        <CopyButton text={secret} size="md" />
      </Card>
    </div>
  );
}

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
        <p role="alert" className="text-2xs text-accent-red">Couldn't load the secret.</p>
      ) : secret ? (
        <SecretCard secret={secret} />
      ) : (
        <p className="text-2xs text-text-muted">Loading the secret…</p>
      )}
    </Modal>
  );
}
