import { useMutation } from "@tanstack/react-query";
import {
  createInstanceApiKeyMutation,
  deleteInstanceApiKeyMutation,
} from "@/client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates an instance API key, refreshing the list. The response carries the only copy of the
 * secret the caller will ever see, so it has to be shown before the mutation's data is discarded.
 */
export function useCreateInstanceApiKey() {
  const invalidate = useInvalidateByIds("listInstanceApiKeys");

  return useMutation({
    ...createInstanceApiKeyMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Revokes an instance API key, refreshing the list. Anything still authenticating with it starts
 * failing at once.
 */
export function useDeleteInstanceApiKey() {
  const invalidate = useInvalidateByIds("listInstanceApiKeys");

  return useMutation({
    ...deleteInstanceApiKeyMutation(),
    onSuccess: invalidate,
  });
}
