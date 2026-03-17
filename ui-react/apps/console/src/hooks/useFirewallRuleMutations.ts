import { useMutation } from "@tanstack/react-query";
import {
  createFirewallRuleMutation,
  updateFirewallRuleMutation,
  deleteFirewallRuleMutation,
} from "../client/@tanstack/react-query.gen";
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
