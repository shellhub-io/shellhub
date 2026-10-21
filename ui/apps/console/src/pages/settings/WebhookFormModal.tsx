import { useController, type Control } from "react-hook-form";
import { useMutation } from "@tanstack/react-query";
import { BoltIcon, CheckIcon } from "@heroicons/react/24/outline";
import { isSdkError } from "@/api/errors";
import { webhookCreateMutation, type Webhook } from "@/client";
import FormModal from "@/components/common/FormModal";
import CheckboxField from "@/components/common/fields/CheckboxField";
import FieldError from "@/components/common/fields/FieldError";
import { FormInputField } from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useUpdateWebhook } from "@/hooks/useWebhookRequests";
import { LABEL } from "@/utils/styles";
import { WEBHOOK_EVENTS, webhookSchema, type WebhookFormValues } from "./webhookSchema";

function EventsField({ control }: { control: Control<WebhookFormValues> }) {
  const {
    field,
    fieldState: { error },
  } = useController({ name: "events", control });

  return (
    <fieldset aria-describedby={error ? "webhook-events-error" : undefined}>
      <legend className={LABEL}>Events</legend>
      <div className="space-y-2">
        {WEBHOOK_EVENTS.map((event) => (
          <CheckboxField
            key={event.value}
            id={`webhook-event-${event.value}`}
            label={event.label}
            checked={field.value.includes(event.value)}
            onChange={(checked) =>
              field.onChange(
                checked
                  ? [...field.value, event.value]
                  : field.value.filter((value) => value !== event.value),
              )
            }
          />
        ))}
      </div>
      <FieldError id="webhook-events-error">{error?.message}</FieldError>
    </fieldset>
  );
}

/**
 * Creates a webhook, or edits one when webhook is set. A create hands its secret to onCreated,
 * since the response is the only place besides reveal that carries it.
 */
export default function WebhookFormModal({
  open,
  webhook,
  onClose,
  onCreated,
}: {
  open: boolean;
  webhook: Webhook | null;
  onClose: () => void;
  onCreated: (secret: string) => void;
}) {
  const invalidate = useInvalidateByIds("webhookList");
  const createWebhook = useMutation({ ...webhookCreateMutation(), onSuccess: invalidate });
  const updateWebhook = useUpdateWebhook();
  const form = useDrawerForm(open, webhookSchema, {
    url: webhook?.url ?? "",
    events: webhook?.events ?? [],
  });
  const { control, setError } = form;

  const onValid = async (values: WebhookFormValues) => {
    try {
      if (webhook) {
        await updateWebhook.mutateAsync({ path: { id: webhook.id }, body: values });
        onClose();
      } else {
        const created = await createWebhook.mutateAsync({ body: values });
        onCreated(created.secret);
      }
    } catch (err) {
      if (isSdkError(err) && err.status === 400) {
        setError("url", { message: "Enter an http or https URL." });
      } else if (isSdkError(err) && err.status === 403) {
        setError("root", {
          message: webhook
            ? "You don't have permission to edit this webhook."
            : "Couldn't add the webhook: the namespace has the most it can hold, or you don't have permission.",
        });
      } else {
        setError("root", { message: "Couldn't save the webhook. Please try again." });
      }
    }
  };

  return (
    <FormModal
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<BoltIcon />}
      title={webhook ? "Edit webhook" : "Add webhook"}
      description="ShellHub sends a signed POST to this URL when a device enters one of the chosen states."
      submitLabel={webhook ? "Save" : "Add webhook"}
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
      requireDirty={!!webhook}
    >
      <FormInputField
        name="url"
        control={control}
        id="webhook-url"
        label="URL"
        placeholder="https://hooks.example.com/shellhub"
      />
      <EventsField control={control} />
    </FormModal>
  );
}
