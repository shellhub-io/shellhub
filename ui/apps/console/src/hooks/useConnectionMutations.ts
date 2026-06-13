import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createConnection,
  updateConnection,
  deleteConnection,
  type ConnectionBody,
} from "@/api/connections";

export function useCreateConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: ConnectionBody) => createConnection(body),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["connections"] }),
  });
}

export function useUpdateConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: ConnectionBody }) =>
      updateConnection(id, body),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["connections"] }),
  });
}

export function useDeleteConnection() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => deleteConnection(id),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["connections"] }),
  });
}
