import { useQuery } from "@tanstack/react-query";
import {
  getNamespaceMembershipInvitationList,
  type GetNamespaceMembershipInvitationListData,
  type MembershipInvitation,
} from "@/client";
import {
  getNamespaceMembershipInvitationListQueryKey,
  resolveInvitationOptions,
} from "@/client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import {
  invitationStatusFilter,
  type InvitationStatus,
} from "@/utils/invitations";

interface UseNamespaceInvitationsParams {
  tenantId: string;
  status?: InvitationStatus;
  page?: number;
  perPage?: number;
  enabled?: boolean;
}

export function useResolveInvitation(invite: string) {
  const { data, isLoading, isError } = useQuery({
    ...resolveInvitationOptions({ query: { invite } }),
    enabled: !!invite,
    retry: false,
    staleTime: Infinity,
  });

  const resolved =
    data?.tenant_id && data.user_id && data.status
      ? {
          tenantId: data.tenant_id,
          userId: data.user_id,
          email: data.email ?? "",
          status: data.status,
        }
      : null;

  return { resolved, isLoading, isError };
}

export function useNamespaceInvitations({
  tenantId,
  status = "pending",
  page = 1,
  perPage = 10,
  enabled = true,
}: UseNamespaceInvitationsParams) {
  const options = {
    path: { tenant: tenantId },
    query: {
      filter: invitationStatusFilter(status),
      page,
      per_page: perPage,
    },
  } satisfies {
    path: GetNamespaceMembershipInvitationListData["path"];
    query: GetNamespaceMembershipInvitationListData["query"];
  };

  const result = useQuery<PaginatedResult<MembershipInvitation>>({
    queryKey: getNamespaceMembershipInvitationListQueryKey(options),
    queryFn: paginatedQueryFn(getNamespaceMembershipInvitationList, options),
    enabled: enabled && !!tenantId,
  });

  return {
    invitations: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
