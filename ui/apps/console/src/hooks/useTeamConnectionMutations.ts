import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createTeamConnection,
  updateTeamConnection,
  deleteTeamConnection,
  putTeamConnectionPrefs,
  type TeamConnectionBody,
  type TeamConnectionPrefsBody,
} from "@/api/teamConnections";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

export function useCreateTeamConnection() {
  const invalidate = useInvalidateByIds("listTeamConnections");
  return useMutation({
    mutationFn: (body: TeamConnectionBody) => createTeamConnection(body),
    onSuccess: () => invalidate(),
  });
}

export function useUpdateTeamConnection() {
  const queryClient = useQueryClient();
  const invalidate = useInvalidateByIds("listTeamConnections");
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: TeamConnectionBody }) =>
      updateTeamConnection(id, body),
    onSuccess: (_d, { id }) => {
      void invalidate();
      // The target may have moved, so the cached reachability is stale.
      void queryClient.invalidateQueries({
        queryKey: ["team-connection-status", id],
      });
    },
  });
}

export function useDeleteTeamConnection() {
  const invalidate = useInvalidateByIds("listTeamConnections");
  return useMutation({
    mutationFn: (id: string) => deleteTeamConnection(id),
    onSuccess: () => invalidate(),
  });
}

export function useUpdateTeamConnectionPrefs() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: TeamConnectionPrefsBody }) =>
      putTeamConnectionPrefs(id, body),
    onSuccess: (_data, { id }) =>
      queryClient.invalidateQueries({
        queryKey: ["team-connection-prefs", id],
      }),
  });
}
