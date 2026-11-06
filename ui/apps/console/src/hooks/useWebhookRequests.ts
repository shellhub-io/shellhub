import { useMutation } from "@tanstack/react-query";
import { webhookUpdateMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * Changes a webhook's URL, event types or enabled flag.
 */
export function useUpdateWebhook() {
  const invalidate = useInvalidateByIds("webhookList");
  return useMutation({ ...webhookUpdateMutation(), onSuccess: invalidate });
}
