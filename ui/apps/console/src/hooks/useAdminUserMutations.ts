import { useMutation } from "@tanstack/react-query";
import {
  createUserAdminMutation,
  adminUpdateUserMutation,
  adminDeleteUserMutation,
  adminResetUserPasswordMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateUser() {
  const invalidate = useInvalidateByIds("getUsers");
  return useMutation({
    ...createUserAdminMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateUser() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminUpdateUserMutation(),
    onSuccess: invalidate,
  });
}

export function useDeleteUser() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminDeleteUserMutation(),
    onSuccess: invalidate,
  });
}

export function useResetUserPassword() {
  const invalidate = useInvalidateByIds("getUsers", "getUser");
  return useMutation({
    ...adminResetUserPasswordMutation(),
    onSuccess: invalidate,
  });
}
