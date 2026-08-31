import { useMutation } from "@tanstack/react-query";
import {
  createUserAdminMutation,
  adminUpdateUserMutation,
  adminDeleteUserMutation,
  adminResetUserPasswordMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates a user as an admin, refreshing the user list.
 */
export function useCreateUser() {
  const invalidate = useInvalidateByIds("getUsers");
  return useMutation({
    ...createUserAdminMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates a user as an admin, refreshing the list and the user's own query.
 */
export function useUpdateUser() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminUpdateUserMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Deletes a user as an admin, refreshing the list and the user's own query.
 */
export function useDeleteUser() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminDeleteUserMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Resets a user's password as an admin. The user is not notified by this call, so whoever ran it
 * has to tell them.
 */
export function useResetUserPassword() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminResetUserPasswordMutation(),
    onSuccess: invalidate,
  });
}
