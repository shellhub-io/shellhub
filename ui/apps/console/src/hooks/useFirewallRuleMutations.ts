import { useMutation } from "@tanstack/react-query";
import {
  createFirewallRuleMutation,
  updateFirewallRuleMutation,
  deleteFirewallRuleMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...createFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...updateFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...deleteFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}
