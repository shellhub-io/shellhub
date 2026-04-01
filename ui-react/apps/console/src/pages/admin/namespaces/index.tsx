import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import {
  ServerStackIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  TrashIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAdminNamespaces } from "../../../hooks/useAdminNamespaces";
import type { Namespace } from "../../../client";
import PageHeader from "../../../components/common/PageHeader";
import Pagination from "../../../components/common/Pagination";
import EditNamespaceDrawer from "./EditNamespaceDrawer";
import DeleteNamespaceDialog from "./DeleteNamespaceDialog";
import { TH as TH_BASE } from "../../../utils/styles";
import { formatDateShort } from "../../../utils/date";
import { formatMaxDevices } from "./utils";

const TH = `${TH_BASE} whitespace-nowrap`;
const PER_PAGE = 10;
const SEARCH_DEBOUNCE_MS = 300;

function getOwnerEmail(namespace: Namespace): string {
  const owner = namespace.members?.find((m) => m.id === namespace.owner);
  return owner?.email || namespace.owner;
}

export default function AdminNamespaces() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [searchInput, setSearchInput] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useState("");
  const [editTarget, setEditTarget] = useState<Namespace | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Namespace | null>(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchInput);
      setPage(1);
    }, SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(timer);
  }, [searchInput]);

  const { namespaces, totalCount, isLoading, error } = useAdminNamespaces({
    page,
    perPage: PER_PAGE,
    search: debouncedSearch,
  });

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  return (
    <div>
      <PageHeader
        icon={<ServerStackIcon className="w-6 h-6" />}
        overline="Namespace Management"
        title="Namespaces"
        description="Manage all namespaces in the instance"
      />

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
          placeholder="Search by name..."
          aria-label="Search namespaces by name"
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
                <th className={TH}>Owner</th>
                <th className={TH}>Devices</th>
                <th className={TH}>Max Devices</th>
                <th className={TH}>Created</th>
                <th className={`${TH} text-right`}>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {isLoading && namespaces.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-16 text-center">
                    <div
                      className="flex items-center justify-center gap-3"
                      role="status"
                    >
                      <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                      <span className="text-xs font-mono text-text-muted">
                        Loading namespaces...
                      </span>
                    </div>
                  </td>
                </tr>
              ) : namespaces.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-16 text-center">
                    <ServerStackIcon
                      className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
                      strokeWidth={1}
                    />
                    <p className="text-xs font-mono text-text-muted">
                      {debouncedSearch
                        ? `No namespaces matching "${debouncedSearch}"`
                        : "No namespaces found"}
                    </p>
                  </td>
                </tr>
              ) : (
                namespaces.map((ns) => (
                  <tr
                    key={ns.tenant_id}
                    onClick={() =>
                      void navigate(`/admin/namespaces/${ns.tenant_id}`)}
                    className="group hover:bg-hover-subtle transition-colors cursor-pointer"
                  >
                    <td className="px-4 py-3.5">
                      <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
                        {ns.name}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {getOwnerEmail(ns)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {ns.devices_accepted_count}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {formatMaxDevices(ns.max_devices)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5">
                      <span className="text-xs text-text-secondary">
                        {formatDateShort(ns.created_at)}
                      </span>
                    </td>
                    <td className="px-4 py-3.5 text-right">
                      <div className="flex items-center justify-end gap-1">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setEditTarget(ns);
                          }}
                          className="p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
                          title="Edit namespace"
                          aria-label={`Edit ${ns.name}`}
                        >
                          <PencilSquareIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            setDeleteTarget(ns);
                          }}
                          className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
                          title="Delete namespace"
                          aria-label={`Delete ${ns.name}`}
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
        itemLabel="namespace"
        onPageChange={setPage}
      />

      <EditNamespaceDrawer
        open={!!editTarget}
        onClose={() => setEditTarget(null)}
        namespace={editTarget}
      />

      <DeleteNamespaceDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        namespace={deleteTarget}
      />
    </div>
  );
}
