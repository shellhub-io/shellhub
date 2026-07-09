import { useState, type MouseEvent } from "react";
import {
  UserPlusIcon,
  CheckIcon,
  XMarkIcon,
} from "@heroicons/react/24/outline";
import { useAdminAccountRequests } from "@/hooks/useAdminAccountRequests";
import {
  useApproveAccountRequest,
  useRejectAccountRequest,
} from "@/hooks/useAdminAccountRequestMutations";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import type { UserAdminResponse } from "@/client";
import DataTable, { type Column } from "@/components/common/DataTable";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { Callout, IconButton } from "@shellhub/design-system/primitives";

const PER_PAGE = 10;

type AccountRequestsParams = { page: number };

const DEFAULTS: AccountRequestsParams = { page: 1 };

export default function AccountRequestsTab() {
  const { params, setPage } = usePaginatedListState<AccountRequestsParams>({
    prefix: "req",
    defaults: DEFAULTS,
  });
  const [approveTarget, setApproveTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const [rejectTarget, setRejectTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const [error, setError] = useState("");

  const {
    requests,
    totalCount,
    isLoading,
    error: loadError,
  } = useAdminAccountRequests({ page: params.page, perPage: PER_PAGE });
  const approve = useApproveAccountRequest();
  const reject = useRejectAccountRequest();

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<UserAdminResponse>[] = [
    {
      key: "name",
      header: "Name",
      render: (u) => (
        <span className="text-sm font-medium text-text-primary">{u.name}</span>
      ),
    },
    {
      key: "email",
      header: "Email",
      render: (u) => (
        <span className="text-xs text-text-secondary">{u.email}</span>
      ),
    },
    {
      key: "username",
      header: "Username",
      render: (u) => (
        <code className="text-2xs font-mono text-text-muted">{u.username}</code>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (u) => (
        <div className="flex items-center justify-end gap-1">
          <IconButton
            variant="primary"
            title="Approve account"
            aria-label={`Approve account for ${u.email}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setApproveTarget(u);
            }}
          >
            <CheckIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            variant="danger"
            title="Reject account"
            aria-label={`Reject account for ${u.email}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setRejectTarget(u);
            }}
          >
            <XMarkIcon className="w-4 h-4" />
          </IconButton>
        </div>
      ),
    },
  ];

  return (
    <div className="animate-fade-in">
      <p className="text-sm text-text-muted mb-5">
        Accounts a namespace admin provisioned, awaiting your approval.
        Approving lets an activation link be issued from the namespace's members
        list.
      </p>

      {loadError && (
        <Callout variant="error" className="mb-4">
          {loadError.message}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={requests}
        rowKey={(u) => u.id}
        isLoading={isLoading}
        loadingMessage="Loading requests..."
        page={params.page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="request"
        onPageChange={setPage}
        emptyState={
          <div className="text-center">
            <UserPlusIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              No accounts awaiting approval
            </p>
          </div>
        }
      />

      <ConfirmDialog
        open={!!approveTarget}
        onClose={() => {
          setError("");
          setApproveTarget(null);
        }}
        onConfirm={async () => {
          if (!approveTarget) return;
          setError("");
          try {
            await approve.mutateAsync({ path: { id: approveTarget.id } });
            setApproveTarget(null);
          } catch {
            setError("Failed to approve the account. Please try again.");
          }
        }}
        title="Approve Account"
        description={
          <>
            Approve the account for{" "}
            <span className="font-medium text-text-primary">
              {approveTarget?.email}
            </span>
            ? An activation link can then be issued from the members list.
            {error && (
              <span className="block mt-2 text-accent-red text-2xs">
                {error}
              </span>
            )}
          </>
        }
        confirmLabel="Approve"
        variant="primary"
      />

      <ConfirmDialog
        open={!!rejectTarget}
        onClose={() => {
          setError("");
          setRejectTarget(null);
        }}
        onConfirm={async () => {
          if (!rejectTarget) return;
          setError("");
          try {
            await reject.mutateAsync({ path: { id: rejectTarget.id } });
            setRejectTarget(null);
          } catch {
            setError("Failed to reject the account. Please try again.");
          }
        }}
        title="Reject Account"
        description={
          <>
            Reject and delete the provisioned account for{" "}
            <span className="font-medium text-text-primary">
              {rejectTarget?.email}
            </span>
            ? This cannot be undone.
            {error && (
              <span className="block mt-2 text-accent-red text-2xs">
                {error}
              </span>
            )}
          </>
        }
        confirmLabel="Reject"
      />
    </div>
  );
}
