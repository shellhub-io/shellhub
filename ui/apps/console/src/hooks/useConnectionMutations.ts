import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createConnection,
  updateConnection,
  deleteConnection,
  type ConnectionBody,
} from "@/api/connections";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

export function useCreateConnection() {
  const invalidate = useInvalidateByIds("listConnections");
  return useMutation({
    mutationFn: (body: ConnectionBody) => createConnection(body),
    onSuccess: () => invalidate(),
  });
}

export function useUpdateConnection() {
  const queryClient = useQueryClient();
  const invalidate = useInvalidateByIds("listConnections");
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: ConnectionBody }) =>
      updateConnection(id, body),
    onSuccess: (_d, { id }) => {
      void invalidate();
      // The target may have moved, so the cached reachability is stale.
      void queryClient.invalidateQueries({
        queryKey: ["connection-status", id],
      });
    },
  });
}

export function useDeleteConnection() {
  const invalidate = useInvalidateByIds("listConnections");
  return useMutation({
    mutationFn: (id: string) => deleteConnection(id),
    onSuccess: () => invalidate(),
  });
}
