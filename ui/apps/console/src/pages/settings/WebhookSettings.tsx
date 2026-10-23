import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { PlusIcon } from "@heroicons/react/24/outline";
import { Badge, Button, Card } from "@shellhub/design-system/primitives";
import { webhookList, webhookListQueryKey, type Webhook } from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";
import PageLoader from "@/components/common/PageLoader";
import FieldError from "@/components/common/fields/FieldError";
import SettingsSection from "@/components/settings/SettingsSection";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useUpdateWebhook } from "@/hooks/useWebhookRequests";
import DeleteWebhookDialog from "./DeleteWebhookDialog";
import WebhookDeliveries from "./WebhookDeliveries";
import WebhookFormModal from "./WebhookFormModal";
import WebhookSecretModal from "./WebhookSecretModal";

function WebhookCard({
  webhook,
  canEdit,
  onEdit,
  onReveal,
  onDelete,
}: {
  webhook: Webhook;
  canEdit: boolean;
  onEdit: () => void;
  onReveal?: () => void;
  onDelete?: () => void;
}) {
  const updateWebhook = useUpdateWebhook();
  const updateErrorId = `webhook-${webhook.id}-update-error`;
  const [showDeliveries, setShowDeliveries] = useState(false);

  return (
    <Card as="article" aria-label={webhook.url} className="rounded-xl p-4 space-y-3">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
        <div className="min-w-0 flex-1 space-y-2">
          <p className="font-mono text-sm text-text-primary break-all">{webhook.url}</p>
          <div className="flex flex-wrap gap-1.5">
            {webhook.events.map((event) => (
              <Badge key={event} color="cyan">
                {event}
              </Badge>
            ))}
            <Badge shape="pill" color={webhook.enabled ? "green" : "yellow"}>
              {webhook.enabled ? "Enabled" : "Paused"}
            </Badge>
          </div>
        </div>
        <div className="flex shrink-0 flex-wrap gap-2">
          {canEdit && (
            <>
              <Button
                size="sm"
                variant="ghost"
                loading={updateWebhook.isPending}
                aria-describedby={updateWebhook.isError ? updateErrorId : undefined}
                onClick={() =>
                  updateWebhook.mutate({
                    path: { id: webhook.id },
                    body: { enabled: !webhook.enabled },
                  })
                }
              >
                {webhook.enabled ? "Pause" : "Resume"}
              </Button>
              <Button size="sm" variant="ghost" onClick={onEdit}>
                Edit
              </Button>
            </>
          )}
          {onReveal && (
            <Button size="sm" variant="ghost" onClick={onReveal}>
              Show secret
            </Button>
          )}
          {onDelete && (
            <Button size="sm" variant="dangerSoft" onClick={onDelete}>
              Delete
            </Button>
          )}
        </div>
      </div>
      <FieldError id={updateErrorId} role="alert">
        {updateWebhook.isError && "Couldn't update the webhook. Try again."}
      </FieldError>
      <Button
        size="sm"
        variant="outline"
        aria-expanded={showDeliveries}
        onClick={() => setShowDeliveries((shown) => !shown)}
      >
        {showDeliveries ? "Hide deliveries" : "Show deliveries"}
      </Button>
      {showDeliveries && <WebhookDeliveries webhookId={webhook.id} />}
    </Card>
  );
}

/**
 * The namespace's webhooks: where its events are sent, which ones, and how the recent
 * deliveries went.
 */
export default function WebhookSettings() {
  const { data, isLoading, error } = useQuery(
    paginatedQueryOptions(webhookList, webhookListQueryKey, {
      query: { page: 1, per_page: 100 },
    }),
  );
  const webhooks = data?.data ?? [];
  const canCreate = useHasPermission("webhook:create");
  const canEdit = useHasPermission("webhook:edit");
  const canDelete = useHasPermission("webhook:delete");
  const canReveal = useHasPermission("webhook:reveal");

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState<Webhook | null>(null);
  const [createdSecret, setCreatedSecret] = useState("");
  const [revealing, setRevealing] = useState<string | null>(null);
  const [deleting, setDeleting] = useState<Webhook | null>(null);

  const openForm = (webhook: Webhook | null) => {
    setEditing(webhook);
    setFormOpen(true);
  };

  if (isLoading) return <PageLoader label="Loading webhooks" padding="lg" />;

  return (
    <SettingsSection
      title="Webhooks"
      description="Get a signed HTTP request when something happens in this namespace."
      action={
        canCreate && (
          <Button
            variant="primary"
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={() => openForm(null)}
          >
            Add webhook
          </Button>
        )
      }
    >
      {error ? (
        <p role="alert" className="text-sm text-accent-red">Couldn't load the webhooks.</p>
      ) : webhooks.length === 0 ? (
        <p className="text-sm text-text-muted">
          No webhooks yet. Add one to stop polling the device list.
        </p>
      ) : (
        webhooks.map((webhook) => (
          <WebhookCard
            key={webhook.id}
            webhook={webhook}
            canEdit={canEdit}
            onEdit={() => openForm(webhook)}
            onReveal={canReveal ? () => setRevealing(webhook.id) : undefined}
            onDelete={canDelete ? () => setDeleting(webhook) : undefined}
          />
        ))
      )}

      <WebhookFormModal
        open={formOpen}
        webhook={editing}
        onClose={() => setFormOpen(false)}
        onCreated={(secret) => {
          setFormOpen(false);
          setCreatedSecret(secret);
        }}
      />
      <WebhookSecretModal
        created={createdSecret || undefined}
        webhookId={revealing}
        onClose={() => {
          setCreatedSecret("");
          setRevealing(null);
        }}
      />
      <DeleteWebhookDialog webhook={deleting} onClose={() => setDeleting(null)} />
    </SettingsSection>
  );
}
