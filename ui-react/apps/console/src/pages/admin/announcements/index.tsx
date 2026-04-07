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
import Pagination from "@/components/common/Pagination";
import DeleteAnnouncementDialog from "./DeleteAnnouncementDialog";
import { TH as TH_BASE } from "@/utils/styles";
import { formatDateShort } from "@/utils/date";
import type { AnnouncementShort } from "@/client";

const TH = `${TH_BASE} whitespace-nowrap`;
const PER_PAGE = 10;

function AnnouncementRow({
  announcement,
  onEdit,
  onDelete,
}: {
  announcement: AnnouncementShort;
  onEdit: () => void;
  onDelete: () => void;
}) {
  const navigate = useNavigate();

  return (
    <tr
      onClick={() => void navigate(`/admin/announcements/${announcement.uuid}`)}
      className="group hover:bg-hover-subtle transition-colors cursor-pointer"
    >
      {/* UUID */}
      <td className="px-4 py-3.5">
        <span className="inline-flex items-center px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-mono font-medium max-w-[120px] truncate">
          {announcement.uuid.slice(0, 8)}
        </span>
      </td>

      {/* Title */}
      <td className="px-4 py-3.5">
        <span className="text-sm text-text-primary truncate block max-w-[400px]">
          {announcement.title}
        </span>
      </td>

      {/* Date */}
      <td className="px-4 py-3.5">
        <span className="text-xs text-text-secondary font-mono">
          {formatDateShort(announcement.date)}
        </span>
      </td>

      {/* Actions */}
      <td className="px-4 py-3.5 text-right">
        <div className="flex items-center justify-end gap-1">
          <button
            onClick={(e) => {
              e.stopPropagation();
              onEdit();
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/5 transition-colors"
            title="Edit announcement"
            aria-label={`Edit ${announcement.title}`}
          >
            <PencilSquareIcon className="w-4 h-4" />
          </button>
          <button
            onClick={(e) => {
              e.stopPropagation();
              onDelete();
            }}
            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/5 transition-colors"
            title="Delete announcement"
            aria-label={`Delete ${announcement.title}`}
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </td>
    </tr>
  );
}

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

      {/* Table */}
      <div className="bg-card border border-border rounded-xl overflow-hidden animate-fade-in">
        <div className="overflow-x-auto">
          <table className="w-full" aria-label="Announcements">
            <thead>
              <tr className="border-b border-border bg-surface/50">
                <th className={TH}>UUID</th>
                <th className={TH}>Title</th>
                <th className={TH}>Date</th>
                <th className={`${TH} text-right`}>Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {isLoading && announcements.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-4 py-16 text-center">
                    <div
                      className="flex items-center justify-center gap-3"
                      role="status"
                    >
                      <span className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
                      <span className="text-xs font-mono text-text-muted">
                        Loading announcements...
                      </span>
                    </div>
                  </td>
                </tr>
              ) : announcements.length === 0 ? (
                <tr>
                  <td colSpan={4} className="px-4 py-16 text-center">
                    <MegaphoneIcon
                      className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
                      strokeWidth={1}
                    />
                    <p className="text-xs font-mono text-text-muted">
                      No announcements found
                    </p>
                  </td>
                </tr>
              ) : (
                announcements.map((a) => (
                  <AnnouncementRow
                    key={a.uuid}
                    announcement={a}
                    onEdit={() =>
                      void navigate(`/admin/announcements/${a.uuid}/edit`)
                    }
                    onDelete={() => setDeleteTarget(a)}
                  />
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
        itemLabel="announcement"
        onPageChange={setPage}
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
