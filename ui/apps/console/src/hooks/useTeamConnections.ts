import { useQuery } from "@tanstack/react-query";
import { getConfig } from "@/env";
import {
  listTeamConnections as listTeamConnectionsSdk,
  type TeamConnection as GeneratedTeamConnection,
} from "@/client";
import { listTeamConnectionsQueryKey } from "@/client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import { isSdkError } from "@/api/errors";
import {
  getTeamConnectionStatus,
  getTeamConnectionPrefs,
} from "@/api/teamConnections";

// Team connections only exist on Cloud/Enterprise editions. On community the
// endpoints aren't served, so the queries stay disabled.
function teamEnabled(): boolean {
  const cfg = getConfig();
  return !!cfg.cloud || !!cfg.enterprise;
}

export function useTeamConnections({
  page = 1,
  perPage = 100,
}: { page?: number; perPage?: number } = {}) {
  const options = { query: { page, per_page: perPage } };

  const result = useQuery<PaginatedResult<GeneratedTeamConnection>>({
    queryKey: listTeamConnectionsQueryKey(options),
    queryFn: async () => {
      try {
        return await paginatedQueryFn(listTeamConnectionsSdk, options)();
      } catch (err) {
        // An edition without the feature replies 402; degrade to empty so the
        // page hides the team UI instead of erroring.
        if (isSdkError(err) && err.status === 402) {
          return { data: [], totalCount: 0 };
        }

        throw err;
      }
    },
    enabled: teamEnabled(),
  });

  return {
    teamConnections: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}

export function useTeamConnectionStatus(id: string) {
  const result = useQuery({
    queryKey: ["team-connection-status", id],
    queryFn: () => getTeamConnectionStatus(id),
    enabled: teamEnabled() && !!id,
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  return {
    online: result.isLoading ? null : (result.data ?? false),
    isChecking: result.isLoading,
  };
}

// The caller's own auth preference for a team connection.
export function useTeamConnectionPrefs(id: string | undefined) {
  const result = useQuery({
    queryKey: ["team-connection-prefs", id],
    queryFn: () => getTeamConnectionPrefs(id as string),
    enabled: teamEnabled() && !!id,
  });

  return {
    prefs: result.data ?? null,
    isLoading: result.isLoading,
  };
}
