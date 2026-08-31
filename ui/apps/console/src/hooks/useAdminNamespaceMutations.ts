import { useMutation } from "@tanstack/react-query";
import {
  editNamespaceAdminMutation,
  deleteNamespaceAdminMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Edits a namespace as an admin, refreshing both the list and the detail query.
 */
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

/**
 * Deletes a namespace as an admin. Refreshes the list and the detail query; the namespace and
 * everything in it are gone, so there is nothing to undo.
 */
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
