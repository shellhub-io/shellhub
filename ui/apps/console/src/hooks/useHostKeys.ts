import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  getHostKey,
  acceptHostKey,
  forgetHostKey,
  type HostKeyScope,
} from "@/api/hostKeys";

export function useHostKey(
  host: string | undefined,
  port: number | undefined,
  scope: HostKeyScope,
  enabled = true,
) {
  const result = useQuery({
    queryKey: ["host-key", scope, host, port],
    queryFn: () => getHostKey(host as string, port as number, scope),
    enabled: enabled && !!host && !!port,
  });

  return {
    knownHost: result.data ?? null,
    isLoading: result.isLoading,
    refetch: result.refetch,
  };
}

export function useAcceptHostKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: acceptHostKey,
    onSuccess: (_d, vars) =>
      queryClient.invalidateQueries({
        queryKey: ["host-key", vars.scope, vars.host, vars.port],
      }),
  });
}

export function useForgetHostKey() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({
      host,
      port,
      scope,
    }: {
      host: string;
      port: number;
      scope: HostKeyScope;
    }) => forgetHostKey(host, port, scope),
    onSuccess: (_d, vars) =>
      queryClient.invalidateQueries({
        queryKey: ["host-key", vars.scope, vars.host, vars.port],
      }),
  });
}
