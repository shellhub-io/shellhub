import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  MegaphoneIcon,
  ExclamationCircleIcon,
  TrashIcon,
  PencilSquareIcon,
  PlusIcon,
} from "@heroicons/react/24/outline";
import { useAdminAnnouncements } from "@/hooks/useAdminAnnouncements";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import DeleteAnnouncementDialog from "./DeleteAnnouncementDialog";
import { formatDateShort } from "@/utils/date";
import type { AnnouncementShort } from "@/client";

const PER_PAGE = 10;

export default function AdminAnnouncements() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [deleteTarget, setDeleteTarget] = useState<AnnouncementShort | null>(
    null,
  );

  const { announcements, totalCount, isLoading, error } = useAdminAnnouncements(
    {
      page,
      perPage: PER_PAGE,
    },
  );

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<AnnouncementShort>[] = [
    {
      key: "uuid",
      header: "UUID",
      render: (a) => (
        <span className="inline-flex items-center px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-mono font-medium max-w-[120px] truncate">
          {a.uuid.slice(0, 8)}
        </span>
      ),
    },
    {
      key: "title",
      header: "Title",
      render: (a) => (
        <span className="text-sm text-text-primary truncate block max-w-[400px]">
          {a.title}
        </span>
      ),
    },
    {
      key: "date",
      header: "Date",
      render: (a) => (
        <span className="text-xs text-text-secondary font-mono">
          {formatDateShort(a.date)}
        </span>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      headerClassName: "text-right",
      render: (a) => (
        <div className="flex items-center justify-end gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation();
              void navigate(`/admin/announcements/${a.uuid}/edit`);
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/5 transition-colors"
            title="Edit announcement"
            aria-label={`Edit ${a.title}`}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              setDeleteTarget(a);
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
            title="Delete announcement"
            aria-label={`Delete ${a.title}`}
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<MegaphoneIcon className="w-6 h-6" />}
        overline="Instance Administration"
        title="Announcements"
        description="Manage system-wide announcements for all users"
      >
        <button
          onClick={() => void navigate("/admin/announcements/new")}
          className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors"
        >
          <PlusIcon className="w-4 h-4" />
          New
        </button>
      </PageHeader>

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

      <DataTable
        columns={columns}
        data={announcements}
        rowKey={(a) => a.uuid}
        label="Announcements"
        isLoading={isLoading}
        loadingMessage="Loading announcements..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="announcement"
        onPageChange={setPage}
        onRowClick={(a) => void navigate(`/admin/announcements/${a.uuid}`)}
        emptyState={
          <div className="text-center">
            <MegaphoneIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              No announcements found
            </p>
          </div>
        }
      />

      <DeleteAnnouncementDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        announcement={deleteTarget}
        onDeleted={() => {
          if (announcements.length <= 1 && page > 1) {
            setPage(page - 1);
          }
        }}
      />
    </div>
  );
}
