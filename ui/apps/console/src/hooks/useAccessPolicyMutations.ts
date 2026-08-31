import { useMutation } from "@tanstack/react-query";
import {
  createAccessPolicyMutation,
  updateAccessPolicyMutation,
  deleteAccessPolicyMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates an access policy, refreshing the list on success.
 */
export function useCreateAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...createAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates an access policy, refreshing the list on success.
 */
export function useUpdateAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...updateAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes an access policy, refreshing the list on success.
 */
export function useDeleteAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...deleteAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}
