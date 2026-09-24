import { useMutation } from "@tanstack/react-query";
import {
  provisioningKeyCreateMutation,
  provisioningKeyUpdateMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a provisioning key. The response carries the only copy of the key, so it has to be shown
 * before the mutation's data is discarded.
 */
export function useCreateProvisioningKey() {
  const invalidate = useInvalidateByIds("provisioningKeyList");
  return useMutation({
    ...provisioningKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates a provisioning key's name or limits. The key itself is not re-issued.
 */
export function useUpdateProvisioningKey() {
  const invalidate = useInvalidateByIds("provisioningKeyList");
  return useMutation({
    ...provisioningKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}
