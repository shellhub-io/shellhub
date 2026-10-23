import { z } from "zod";
import type { WebhookEventType } from "@/client";
import { zWebhookEventType } from "@/client/zod.gen";

const WEBHOOK_EVENT_LABELS = {
  "device.pending": "Device pending",
  "device.accepted": "Device accepted",
  "device.rejected": "Device rejected",
  "device.removed": "Device removed",
} satisfies Record<WebhookEventType, string>;

/**
 * Every event a webhook can subscribe to, with the label the form shows for it. An event type the
 * API adds fails the type check until it gets a label here.
 */
export const WEBHOOK_EVENTS = zWebhookEventType.options.map((value) => ({
  value,
  label: WEBHOOK_EVENT_LABELS[value],
}));

/**
 * Validates the webhook form: an http or https URL and at least one event.
 */
export const webhookSchema = z.object({
  url: z
    .string()
    .trim()
    .url("Enter an http or https URL.")
    .regex(/^https?:\/\//, "Enter an http or https URL."),
  events: z.array(zWebhookEventType).min(1, "Pick at least one event."),
});

/**
 * The webhook form's values, derived from the schema.
 */
export type WebhookFormValues = z.infer<typeof webhookSchema>;
