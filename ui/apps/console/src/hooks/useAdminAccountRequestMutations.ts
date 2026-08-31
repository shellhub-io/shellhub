import { useMutation } from "@tanstack/react-query";
import { approveUserMutation } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Approves a pending account request. This clears awaiting_approval only — the account stays
 * unconfirmed until the person activates it, which is what lets an activation link be minted
 * for them afterwards from the members list.
 */
export function useApproveAccountRequest() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...approveUserMutation(),
    onSuccess: invalidate,
  });
}
