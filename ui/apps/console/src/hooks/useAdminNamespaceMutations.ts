import { useMutation } from "@tanstack/react-query";
import {
  editNamespaceAdminMutation,
  deleteNamespaceAdminMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAdminEditNamespace() {
  const invalidate = useInvalidateByIds(
    "getNamespacesAdmin",
    "getNamespaceAdmin",
  );
  return useMutation({
    ...editNamespaceAdminMutation(),
    onSuccess: invalidate,
  });
}

export function useAdminDeleteNamespace() {
  const invalidate = useInvalidateByIds(
    "getNamespacesAdmin",
    "getNamespaceAdmin",
  );
  return useMutation({
    ...deleteNamespaceAdminMutation(),
    onSuccess: invalidate,
  });
}
