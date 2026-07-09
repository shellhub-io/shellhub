import { useState, type MouseEvent } from "react";
import { useNavigate } from "react-router-dom";
import {
  UsersIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
} from "@heroicons/react/24/outline";
import { useAdminUsers } from "@/hooks/useAdminUsers";
import { useAdminAccountRequests } from "@/hooks/useAdminAccountRequests";
import { useLoginAsUser } from "@/hooks/useLoginAsUser";
import { useDebouncedValue } from "@/hooks/useDebouncedValue";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { getConfig } from "@/env";
import type { UserAdminResponse } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import UserStatusChip from "./UserStatusChip";
import CreateUserDrawer from "./CreateUserDrawer";
import EditUserDrawer from "./EditUserDrawer";
import DeleteUserDialog from "./DeleteUserDialog";
import AccountRequestsTab from "./AccountRequestsTab";
import {
  Badge,
  Button,
  Callout,
  IconButton,
} from "@shellhub/design-system/primitives";

const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

type AdminUsersParams = {
  page: number;
  search: string;
};

const DEFAULTS: AdminUsersParams = {
  page: 1,
  search: "",
};

export default function AdminUsers() {
  const navigate = useNavigate();
  // Account provisioning requests are an enterprise-only queue; on Cloud a new
  // member self-serves via email invitation, so there is nothing to approve.
  const showRequestsTab = getConfig().enterprise && !getConfig().cloud;
  const [tab, setTab] = useState<"users" | "account-requests">("users");
  const { params, setPage, setSearch } =
    usePaginatedListState<AdminUsersParams>({ defaults: DEFAULTS });
  const debouncedSearch = useDebouncedValue(params.search, SEARCH_DEBOUNCE_MS);
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<UserAdminResponse | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const {
    loginAs,
    loadingId: loginAsId,
    errorId: loginAsError,
  } = useLoginAsUser();

  const { users, totalCount, isLoading, error } = useAdminUsers({
    page: params.page,
    perPage: PER_PAGE,
    search: debouncedSearch,
  });

  // Pending-request count drives the tab badge; only queried in enterprise.
  const { totalCount: requestsCount } = useAdminAccountRequests({
    enabled: showRequestsTab,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<UserAdminResponse>[] = [
    {
      key: "name",
      header: "Name",
      render: (user) => (
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
            {user.name}
          </span>
          {user.admin && <Badge color="yellow">Admin</Badge>}
        </div>
      ),
    },
    {
      key: "email",
      header: "Email",
      render: (user) => (
        <span className="text-xs text-text-secondary">{user.email}</span>
      ),
    },
    {
      key: "username",
      header: "Username",
      render: (user) => (
        <code className="text-2xs font-mono text-text-muted">
          {user.username}
        </code>
      ),
    },
    {
      key: "status",
      header: "Status",
      render: (user) => <UserStatusChip status={user.status} />,
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (user) => (
        <div className="flex items-center justify-end gap-1">
          <IconButton
            variant="primary"
            title="Edit user"
            aria-label={`Edit ${user.name}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setEditTarget(user);
            }}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            variant="primary"
            loading={loginAsId === user.id}
            disabled={loginAsId === user.id}
            className={
              loginAsError === user.id
                ? "text-accent-red hover:text-accent-red hover:bg-accent-red/5"
                : undefined
            }
            title={
              loginAsError === user.id
                ? "Login failed \u2014 click to retry"
                : "Login as user"
            }
            aria-label={`Login as ${user.name}`}
            onClick={(e) => {
              e.stopPropagation();
              void loginAs(user.id);
            }}
          >
            <ArrowRightStartOnRectangleIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            variant="danger"
            title="Delete user"
            aria-label={`Delete ${user.name}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setDeleteTarget(user);
            }}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      ),
    },
  ];

  const onUsersTab = !showRequestsTab || tab === "users";

  return (
    <div>
      <PageHeader
        icon={<UsersIcon className="w-6 h-6" />}
        overline="Account Management"
        title="Users"
        description="Manage all user accounts in the instance"
      >
        {onUsersTab && (
          <Button
            onClick={() => setCreateOpen(true)}
            icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Create User
          </Button>
        )}
      </PageHeader>

      {showRequestsTab && (
        <div className="flex items-center h-8 bg-card border border-border rounded-md p-0.5 w-fit mb-6 animate-fade-in">
          {[
            { label: "Users", value: "users" as const },
            {
              label: "Account Requests",
              value: "account-requests" as const,
              count: requestsCount,
            },
          ].map((t) => (
            <button
              type="button"
              key={t.value}
              onClick={() => setTab(t.value)}
              className={`h-full px-3.5 text-xs font-medium rounded transition-all duration-150 flex items-center gap-1.5 ${
                tab === t.value
                  ? "bg-primary/15 text-primary border border-primary/25"
                  : "text-text-muted hover:text-text-secondary border border-transparent"
              }`}
            >
              {t.label}
              {t.count ? <Badge color="yellow">{t.count}</Badge> : null}
            </button>
          ))}
        </div>
      )}

      {onUsersTab ? (
        <>
          <SearchField
            className="mb-5"
            value={params.search}
            onChange={setSearch}
            placeholder="Search by username..."
            aria-label="Search users by username"
          />

          {error && (
            <Callout variant="error" className="mb-4">
              {error.message}
            </Callout>
          )}

          <DataTable
            columns={columns}
            data={users}
            rowKey={(user) => user.id}
            isLoading={isLoading}
            loadingMessage="Loading users..."
            page={params.page}
            totalPages={totalPages}
            totalCount={totalCount}
            itemLabel="user"
            onPageChange={setPage}
            onRowClick={(user) => void navigate(`/admin/users/${user.id}`)}
            emptyState={
              <div className="text-center">
                <UsersIcon
                  className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
                  strokeWidth={1}
                />
                <p className="text-xs font-mono text-text-muted">
                  {debouncedSearch
                    ? `No users matching "${debouncedSearch}"`
                    : "No users found"}
                </p>
              </div>
            }
          />
        </>
      ) : (
        <AccountRequestsTab />
      )}

      <CreateUserDrawer
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      <EditUserDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        user={editTarget}
      />

      <DeleteUserDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        user={deleteTarget}
      />
    </div>
  );
}
