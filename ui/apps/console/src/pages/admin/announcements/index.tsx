import { type MouseEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  MegaphoneIcon,
  TrashIcon,
  PencilSquareIcon,
  PlusIcon,
} from "@heroicons/react/24/outline";
import { useAdminAnnouncements } from "@/hooks/useAdminAnnouncements";
import PageHeader from "@/components/common/PageHeader";
import { DataTable, type Column } from "@shellhub/design-system/components";
import DeleteAnnouncementDialog from "./DeleteAnnouncementDialog";
import { formatDateShort } from "@/utils/date";
import type { AnnouncementShort } from "@/client";
import {
  Badge,
  Button,
  Callout,
  IconButton,
} from "@shellhub/design-system/primitives";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";

const PER_PAGE = 10;

type AdminAnnouncementsParams = {
  page: number;
};

const DEFAULTS: AdminAnnouncementsParams = { page: 1 };

export default function AdminAnnouncements() {
  const navigate = useNavigate();
  const { params, setPage } = usePaginatedListState<AdminAnnouncementsParams>({
    defaults: DEFAULTS,
  });
  const [deleteTarget, setDeleteTarget] = useState<AnnouncementShort | null>(
    null,
  );

  const { announcements, totalCount, isLoading, error } = useAdminAnnouncements(
    {
      page: params.page,
      perPage: PER_PAGE,
    },
  );

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<AnnouncementShort>[] = [
    {
      key: "uuid",
      header: "UUID",
      render: (a) => (
        <Badge color="primary" className="max-w-32 truncate">
          {a.uuid.slice(0, 8)}
        </Badge>
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
          <IconButton
            variant="primary"
            title="Edit announcement"
            aria-label={`Edit ${a.title}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              void navigate(`/admin/announcements/${a.uuid}/edit`);
            }}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            variant="danger"
            title="Delete announcement"
            aria-label={`Delete ${a.title}`}
            onClick={(e: MouseEvent) => {
              e.stopPropagation();
              setDeleteTarget(a);
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
        icon={<MegaphoneIcon className="w-6 h-6" />}
        overline="Instance Administration"
        title="Announcements"
        description="Manage system-wide announcements for all users"
      >
        <Button
          variant="primary"
          onClick={() => void navigate("/admin/announcements/new")}
          icon={<PlusIcon className="w-4 h-4" />}
        >
          New
        </Button>
      </PageHeader>

      {error && (
        <Callout variant="error" className="mb-4">
          {error.message}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={announcements}
        rowKey={(a) => a.uuid}
        label="Announcements"
        isLoading={isLoading}
        loadingMessage="Loading announcements..."
        page={params.page}
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
          if (announcements.length <= 1 && params.page > 1) {
            setPage(params.page - 1);
          }
        }}
      />
    </div>
  );
}
