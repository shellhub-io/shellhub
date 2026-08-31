import { useMutation } from "@tanstack/react-query";
import {
  createFirewallRuleMutation,
  updateFirewallRuleMutation,
  deleteFirewallRuleMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a firewall rule, refreshing the list.
 */
export function useCreateFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...createFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates a firewall rule, refreshing the list.
 */
export function useUpdateFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...updateFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a firewall rule, refreshing the list.
 */
export function useDeleteFirewallRule() {
  const invalidate = useInvalidateByIds("getFirewallRules");
  return useMutation({
    ...deleteFirewallRuleMutation(),
    onSuccess: invalidate,
  });
}
