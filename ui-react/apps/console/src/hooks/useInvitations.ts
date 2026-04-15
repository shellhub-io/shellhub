import { useQuery } from "@tanstack/react-query";
import {
  getMembershipInvitationList,
  getNamespaceMembershipInvitationList,
  type GetMembershipInvitationListData,
  type GetNamespaceMembershipInvitationListData,
  type MembershipInvitation,
} from "@/client";
import {
  getMembershipInvitationListQueryKey,
  getNamespaceMembershipInvitationListQueryKey,
} from "@/client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import {
  invitationStatusFilter,
  type InvitationStatus,
} from "@/utils/invitations";

interface UseUserInvitationsParams {
  status?: InvitationStatus;
  page?: number;
  perPage?: number;
  enabled?: boolean;
}

export function useUserInvitations({
  status = "pending",
  page = 1,
  perPage = 10,
  enabled = true,
}: UseUserInvitationsParams = {}) {
  const options = {
    query: {
      filter: invitationStatusFilter(status),
      page,
      per_page: perPage,
    },
  } satisfies { query: GetMembershipInvitationListData["query"] };

  const result = useQuery<PaginatedResult<MembershipInvitation>>({
    queryKey: getMembershipInvitationListQueryKey(options),
    queryFn: paginatedQueryFn(getMembershipInvitationList, options),
    enabled,
  });

  return {
    invitations: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}

interface UseNamespaceInvitationsParams {
  tenantId: string;
  status?: InvitationStatus;
  page?: number;
  perPage?: number;
  enabled?: boolean;
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
