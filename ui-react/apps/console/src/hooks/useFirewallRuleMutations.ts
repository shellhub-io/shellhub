import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createFirewallRuleMutation,
  updateFirewallRuleMutation,
  deleteFirewallRuleMutation,
} from "../client/@tanstack/react-query.gen";

function useInvalidateFirewallRules() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ predicate: (query) => {
    const key = query.queryKey[0];
    if (typeof key === "object" && key !== null && "_id" in key) {
      return (key as { _id: string })._id === "getFirewallRules";
    }
    return false;
  } });
}

export function useCreateFirewallRule() {
  const invalidate = useInvalidateFirewallRules();
  return useMutation({
    ...createFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateFirewallRule() {
  const invalidate = useInvalidateFirewallRules();
  return useMutation({
    ...updateFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteFirewallRule() {
  const invalidate = useInvalidateFirewallRules();
  return useMutation({
    ...deleteFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}
