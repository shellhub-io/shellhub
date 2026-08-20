import { useMutation } from "@tanstack/react-query";
import {
  createAccessPolicyMutation,
  updateAccessPolicyMutation,
  deleteAccessPolicyMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...createAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...updateAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteAccessPolicy() {
  const invalidate = useInvalidateByIds("listAccessPolicies");
  return useMutation({
    ...deleteAccessPolicyMutation(),
    onSuccess: invalidate,
  });
}
