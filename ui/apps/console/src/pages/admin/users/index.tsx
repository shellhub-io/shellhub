import { useState, type MouseEvent } from "react";
import { useNavigate } from "react-router-dom";
import {
  UsersIcon,
  PlusIcon,
  PencilSquareIcon,
  TrashIcon,
  CheckIcon,
  ArrowRightStartOnRectangleIcon,
} from "@heroicons/react/24/outline";
import FilterTabs, { type FilterTab } from "@/components/common/FilterTabs";
import { isEnterprise } from "@/env";
import {
  adminUsersQueryOptions,
  usersSearchScope,
  type AdminUserSubset,
} from "@/hooks/useAdminUsers";
import { useLoginAsUser } from "@/hooks/useLoginAsUser";
import { usePagedList } from "@/hooks/usePagedList";
import { useAuthStore } from "@/stores/authStore";
import type { ListParamConstraints } from "@/hooks/paginatedListParams";
import type { UserAdminResponse } from "@/client";
import PageHeader from "@/components/common/PageHeader";
import { adminNavSectionTitle } from "@/components/layout/adminNav";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import UserStatusChip from "./UserStatusChip";
import CreateUserModal from "./CreateUserModal";
import EditUserModal from "./EditUserModal";
import DeleteUserDialog from "./DeleteUserDialog";
import ApproveAccountDialog from "./ApproveAccountDialog";
import RejectAccountDialog from "./RejectAccountDialog";
import {
  Badge,
  Button,
  Callout,
  IconButton,
} from "@shellhub/design-system/primitives";
import { apiErrorMessage } from "@/api/errors";
import { PER_PAGE } from "@/utils/pagination";

const subsetTabs: FilterTab<AdminUserSubset>[] = [
  { label: "All", value: "" },
  { label: "Awaiting approval", value: "awaiting_approval" },
  { label: "Admins", value: "admin" },
  { label: "Unconfirmed", value: "not_confirmed" },
];

const CONSTRAINTS: ListParamConstraints<AdminUsersParams> = {
  subset: subsetTabs.map((tab) => tab.value),
};

type AdminUsersParams = {
  page: number;
  search: string;
  subset: AdminUserSubset;
};

const DEFAULTS: AdminUsersParams = {
  page: 1,
  search: "",
  subset: "",
};

/**
 * Every user on the instance.
 */
export default function AdminUsers() {
  const navigate = useNavigate();
  const isAdmin = useAuthStore((s) => s.isAdmin);
  const tabs = subsetTabs.filter(
    (tab) => tab.value !== "awaiting_approval" || isEnterprise(),
  );
  const visibleSubset = (subset: AdminUserSubset) =>
    tabs.some((tab) => tab.value === subset) ? subset : "";
  const {
    params, appliedSearch: searchTerm, setSearch, setFilter, rows: users, isLoading, error,
    pagination,
  } = usePagedList({
    defaults: DEFAULTS,
    constraints: CONSTRAINTS,
    query: ({ page, search, subset }: AdminUsersParams) =>
      adminUsersQueryOptions(
        { page, perPage: PER_PAGE, search, subset: visibleSubset(subset) },
        isAdmin,
      ),
  });
  const [createOpen, setCreateOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<UserAdminResponse | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const [approveTarget, setApproveTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const [rejectTarget, setRejectTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const {
    loginAs,
    loadingId: loginAsId,
    errorId: loginAsError,
  } = useLoginAsUser();

  const subset = visibleSubset(params.subset);
  const searchScope = usersSearchScope(subset);

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
      render: (user) => (
        <UserStatusChip
          status={user.awaiting_approval ? "awaiting_approval" : user.status}
        />
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (user) => (
        <div className="flex items-center justify-end gap-1">
          {user.awaiting_approval ? (
            <IconButton
              variant="primary"
              title="Approve account"
              aria-label={`Approve account for ${user.email}`}
              onClick={(e: MouseEvent) => {
                e.stopPropagation();
                setApproveTarget(user);
              }}
            >
              <CheckIcon className="w-4 h-4" />
            </IconButton>
          ) : (
            <>
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
                    ? "Login failed — click to retry"
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
            </>
          )}
          <IconButton
            variant="danger"
            title={user.awaiting_approval ? "Reject account" : "Delete user"}
            aria-label={
              user.awaiting_approval
                ? `Reject account for ${user.email}`
                : `Delete ${user.name}`
            }
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              if (user.awaiting_approval) setRejectTarget(user);
              else setDeleteTarget(user);
            }}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<UsersIcon className="w-6 h-6" />}
        overline={adminNavSectionTitle("/admin/users")}
        title="Users"
        description="Manage all user accounts in the instance"
      >
        <Button
          onClick={() => setCreateOpen(true)}
          icon={<PlusIcon className="w-4 h-4" strokeWidth={2} />}
        >
          Create User
        </Button>
      </PageHeader>

      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-5">
        <FilterTabs
          tabs={tabs}
          value={subset}
          onChange={(next) => setFilter("subset", next)}
          label="Filter users"
        />

        <SearchField
          value={params.search}
          onChange={setSearch}
          placeholder={`Search by ${searchScope}...`}
          aria-label={`Search users by ${searchScope}`}
        />
      </div>

      {error && (
        <Callout variant="error" className="mb-4">
          {apiErrorMessage(error)}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={users}
        rowKey={(user) => user.id}
        isLoading={isLoading}
        loadingMessage="Loading users..."
        {...pagination}
        itemLabel="user"
        onRowClick={(user) => void navigate(`/admin/users/${user.id}`)}
        emptyState={
          <div className="text-center">
            <UsersIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              {searchTerm
                ? `No users matching "${searchTerm}"`
                : "No users found"}
            </p>
          </div>
        }
      />

      <CreateUserModal
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      <EditUserModal
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        user={editTarget}
      />

      <DeleteUserDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        user={deleteTarget}
      />

      <ApproveAccountDialog
        open={!!approveTarget}
        onClose={() => setApproveTarget(null)}
        user={approveTarget}
      />

      <RejectAccountDialog
        open={!!rejectTarget}
        onClose={() => setRejectTarget(null)}
        user={rejectTarget}
      />
    </div>
  );
}
