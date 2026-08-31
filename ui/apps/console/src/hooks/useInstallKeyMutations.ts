import { useMutation } from "@tanstack/react-query";
import {
  installKeyCreateMutation,
  installKeyUpdateMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates an install key. The response carries the only copy of the key, so it has to be shown
 * before the mutation's data is discarded.
 */
export function useCreateInstallKey() {
  const invalidate = useInvalidateByIds("installKeyList");
  return useMutation({
    ...installKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates an install key's name or limits. The key itself is not re-issued.
 */
export function useUpdateInstallKey() {
  const invalidate = useInvalidateByIds("installKeyList");
  return useMutation({
    ...installKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}
