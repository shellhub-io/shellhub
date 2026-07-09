import { useMutation } from "@tanstack/react-query";
import {
  approveUserMutation,
  adminDeleteUserMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

// Approving clears the awaiting_approval flag; the account stays not-confirmed until the
// person activates it, so an activation link can then be minted from the members list.
export function useApproveAccountRequest() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...approveUserMutation(),
    onSuccess: invalidate,
  });
}

// Rejecting deletes the provisioned account outright.
export function useRejectAccountRequest() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminDeleteUserMutation(),
    onSuccess: invalidate,
  });
}
