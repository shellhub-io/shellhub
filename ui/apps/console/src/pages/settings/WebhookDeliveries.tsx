import { useQuery } from "@tanstack/react-query";
import { Badge, type BadgeColor } from "@shellhub/design-system/primitives";
import {
  webhookDeliveryList,
  webhookDeliveryListQueryKey,
  type WebhookDelivery,
} from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";
import { formatRelative } from "@/utils/date";
import { WEBHOOK_EVENT_LABELS } from "./webhookSchema";

const STATUS_COLOR: Record<WebhookDelivery["status"], BadgeColor> = {
  pending: "yellow",
  delivered: "green",
  failed: "red",
};

/**
 * The newest deliveries to one webhook, each with how its last attempt went.
 */
export default function WebhookDeliveries({ webhookId }: { webhookId: string }) {
  const { data, isLoading, error } = useQuery(
    paginatedQueryOptions(webhookDeliveryList, webhookDeliveryListQueryKey, {
      path: { id: webhookId },
      query: { page: 1, per_page: 20 },
    }),
  );
  const deliveries = data?.data ?? [];

  if (isLoading) return <p className="text-2xs text-text-muted">Loading deliveries…</p>;
  if (error) return <p role="alert" className="text-2xs text-accent-red">Couldn't load the deliveries.</p>;
  if (deliveries.length === 0) {
    return <p className="text-2xs text-text-muted">No deliveries in the last 7 days.</p>;
  }

  return (
    <ul aria-label="Recent deliveries" className="divide-y divide-border">
      {deliveries.map((delivery) => (
        <li key={delivery.id} className="flex flex-wrap items-center gap-x-3 gap-y-1 py-2 text-xs">
          <Badge shape="pill" color={STATUS_COLOR[delivery.status]}>
            {delivery.status}
          </Badge>
          <span className="text-text-primary">{WEBHOOK_EVENT_LABELS[delivery.event_type]}</span>
          <span className="text-text-muted">
            {delivery.attempts} {delivery.attempts === 1 ? "attempt" : "attempts"}
            {delivery.last_status_code ? `, last answered ${delivery.last_status_code}` : ""}
          </span>
          {delivery.last_error && delivery.status !== "delivered" && (
            <span className="text-accent-red">{delivery.last_error}</span>
          )}
          <span className="ml-auto text-text-muted">{formatRelative(delivery.created_at)}</span>
        </li>
      ))}
    </ul>
  );
}
