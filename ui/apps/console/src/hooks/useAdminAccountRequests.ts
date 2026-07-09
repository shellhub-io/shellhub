import { useQuery } from "@tanstack/react-query";
import {
  getUsers as getUsersSdk,
  type GetUsersData,
  type UserAdminResponse,
} from "../client";
import { getUsersQueryKey } from "../client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";
import { toBase64Json } from "@/utils/encoding";

// Accounts a namespace admin provisioned that a system admin has not approved yet
// are just users flagged awaiting_approval; the "requests" queue is that filter.
const AWAITING_APPROVAL_FILTER = toBase64Json([
  {
    type: "property",
    params: { name: "awaiting_approval", operator: "bool", value: true },
  },
]);

interface UseAdminAccountRequestsParams {
  page?: number;
  perPage?: number;
  enabled?: boolean;
}

export function useAdminAccountRequests({
  page = 1,
  perPage = 10,
  enabled = true,
}: UseAdminAccountRequestsParams = {}) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const query: GetUsersData["query"] = {
    page,
    per_page: perPage,
    filter: AWAITING_APPROVAL_FILTER,
  };
  const options = { query };

  const result = useQuery<PaginatedResult<UserAdminResponse>>({
    queryKey: getUsersQueryKey(options),
    queryFn: paginatedQueryFn(getUsersSdk, options),
    enabled: isAdmin && enabled,
    staleTime: 60 * 1000, // 1 minute
    retry: (count, err) =>
      isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    requests: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
