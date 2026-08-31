import { useMutation } from "@tanstack/react-query";
import {
  apiKeyCreateMutation,
  apiKeyUpdateMutation,
  apiKeyDeleteMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates an API key, refreshing the list. The response carries the only copy of the secret the
 * caller will ever see, so it has to be shown before the mutation's data is discarded.
 */
export function useCreateApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates an API key's name or role, refreshing the list. The secret is not re-issued.
 */
export function useUpdateApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes an API key, refreshing the list. Anything still authenticating with it starts failing
 * at once.
 */
export function useDeleteApiKey() {
  const invalidate = useInvalidateByIds("apiKeyList");
  return useMutation({
    ...apiKeyDeleteMutation(),
    onSuccess: invalidate,
  });
}
