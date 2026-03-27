import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  UsersIcon,
  PlusIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAdminUsers } from "../../../hooks/useAdminUsers";
import { useLoginAsUser } from "../../../hooks/useLoginAsUser";
import type { UserAdminResponse } from "../../../client";
import PageHeader from "../../../components/common/PageHeader";
import Pagination from "../../../components/common/Pagination";
import UserStatusChip from "./UserStatusChip";
import CreateUserDrawer from "./CreateUserDrawer";
import EditUserDrawer from "./EditUserDrawer";
import DeleteUserDialog from "./DeleteUserDialog";
import { TH as TH_BASE } from "../../../utils/styles";

const TH = `${TH_BASE} whitespace-nowrap`;
const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

export default function AdminUsers() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
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

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);
      setPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const { users, totalCount, isLoading, error } = useAdminUsers({
    page,
    perPage: PER_PAGE,
    search: debouncedSearch,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  return (
    <div>
      <PageHeader
        icon={<UsersIcon className="w-6 h-6" />}
        overline="Account Management"
        title="Users"
        description="Manage all user accounts in the instance"
      >
        <button
          onClick={() => setCreateOpen(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all"
        >
          <PlusIcon className="w-4 h-4" strokeWidth={2} />
          Create User
        </button>
      </PageHeader>

      {/* Search */}
      <div className="relative h-8 ml-auto mb-5 animate-fade-in">
        <MagnifyingGlassIcon
          className="absolute left-3 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-text-muted"
          strokeWidth={2}
        />
        <input
          type="text"
          value={searchInput}
          onChange={(e) => setSearchInput(e.target.value)}
          placeholder="Search by username..."
          aria-label="Search users by username"
          className="h-full pl-9 pr-3 bg-card border border-border rounded-md text-xs text-text-primary font-mono placeholder:text-text-secondary focus:outline-none focus:border-primary/40 focus:ring-1 focus:ring-primary/15 transition-all duration-200 w-56"
        />
      </div>

      {error && (
        <div
          role="alert"
          className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down"
        >
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error.message}
        </div>
      )}

      {/* Table */}
      <div className="bg-card border border-border rounded-xl overflow-hidden animate-fade-in">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border bg-surface/50">
                <th className={TH}>Name</th>
                <th className={TH}>Email</th>
                <th className={TH}>Username</th>
                <th className={TH}>Status</th>
                <th className={`${TH} text-right`}>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {isLoading && users.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-16 text-center">
                    <div
                      className="flex items-center justify-center gap-3"
                      role="status"
                    >
                      <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                      <span className="text-xs font-mono text-text-muted">
                        Loading users...
                      </span>
                    </div>
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-16 text-center">
                    <UsersIcon
                      className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
                      strokeWidth={1}
                    />
                    <p className="text-xs font-mono text-text-muted">
                      {debouncedSearch
                        ? `No users matching "${debouncedSearch}"`
                        : "No users found"}
                    </p>
                  </td>
                </tr>
              ) : (
                users.map((user) => (
                  <tr
                    key={user.id}
                    onClick={() => void navigate(`/admin/users/${user.id}`)}
                    className="group hover:bg-hover-subtle transition-colors cursor-pointer"
                  >
                    <td className="px-4 py-3.5">
                      <div className="flex items-center gap-2">
                        <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
                          {user.name}
                        </span>
                        {user.admin && (
                          <span className="inline-flex items-center px-1.5 py-0.5 text-2xs font-semibold rounded bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                            Admin
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {user.email}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <code className="text-2xs font-mono text-text-muted">
                        {user.username}
                      </code>
                    </td>
                    <td className="px-4 py-3.5">
                      <UserStatusChip status={user.status} />
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setEditTarget(user);
                          }}
                          className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                          title="Edit user"
                          aria-label={`Edit ${user.name}`}
                        >
                          <PencilSquareIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            void loginAs(user.id);
                          }}
                          disabled={loginAsId === user.id}
                          className={`p-1.5 rounded-md transition-colors disabled:opacity-dim disabled:cursor-not-allowed ${
                            loginAsError === user.id
                              ? "text-accent-red hover:text-accent-red hover:bg-accent-red/5"
                              : "text-text-muted hover:text-primary hover:bg-primary/5"
                          }`}
                          title={
                            loginAsError === user.id
                              ? "Login failed — click to retry"
                              : "Login as user"
                          }
                          aria-label={`Login as ${user.name}`}
                        >
                          {loginAsId === user.id ? (
                            <span
                              aria-hidden="true"
                              className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin block"
                            />
                          ) : (
                            <ArrowRightStartOnRectangleIcon className="w-4 h-4" />
                          )}
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setDeleteTarget(user);
                          }}
                          className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
                          title="Delete user"
                          aria-label={`Delete ${user.name}`}
                        >
                          <TrashIcon className="w-4 h-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </div>

      <Pagination
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="user"
        onPageChange={setPage}
      />

      {/* Create User Drawer */}
      <CreateUserDrawer
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />

      {/* Edit User Drawer */}
      <EditUserDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        user={editTarget}
      />

      {/* Delete Confirmation */}
      <DeleteUserDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        user={deleteTarget}
      />
    </div>
  );
}
