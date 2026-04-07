import { useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import {
  ChevronRightIcon,
  MegaphoneIcon,
  InformationCircleIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAdminAnnouncement } from "@/hooks/useAdminAnnouncements";
import CopyButton from "@/components/common/CopyButton";
import DeleteAnnouncementDialog from "./DeleteAnnouncementDialog";
import AnnouncementContent from "./AnnouncementContent";
import { formatDateFull } from "@/utils/date";

const LABEL =
  "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";

export default function AnnouncementDetails() {
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const {
    data: announcement,
    isLoading,
    error,
  } = useAdminAnnouncement(uuid ?? "");
  const [deleteOpen, setDeleteOpen] = useState(false);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24" role="status">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading announcement details</span>
      </div>
    );
  }

  if (error || !announcement) {
    return (
      <div className="text-center py-24">
        <MegaphoneIcon
          className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
          strokeWidth={1}
        />
        <p className="text-sm text-text-muted mb-2">Announcement not found</p>
        <Link
          to="/admin/announcements"
          className="text-sm text-primary hover:underline"
        >
          Back to announcements
        </Link>
      </div>
    );
  }

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/announcements"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Announcements
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary truncate min-w-0">
          {announcement.title}
        </span>
      </nav>

      {/* Header */}
      <div className="flex items-start justify-between gap-4 mb-8">
        <div className="flex items-start gap-4">
          <div className="w-14 h-14 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
            <MegaphoneIcon className="w-7 h-7 text-primary" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-text-primary">
              {announcement.title}
            </h1>
            <p className="text-xs text-text-muted mt-1 font-mono">
              {formatDateFull(announcement.date)}
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2 shrink-0">
          <Link
            to={`/admin/announcements/${announcement.uuid}/edit`}
            className="flex items-center gap-2 px-4 py-2.5 border border-border rounded-lg text-sm text-text-secondary hover:text-text-primary hover:border-border-light transition-colors"
          >
            <PencilSquareIcon className="w-4 h-4" />
            Edit
          </Link>
          <button
            onClick={() => setDeleteOpen(true)}
            className="p-2.5 border border-border rounded-lg text-text-muted hover:text-accent-red hover:border-accent-red/30 transition-colors"
            aria-label="Delete announcement"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>

      {/* Properties Card */}
      <div className="bg-card border border-border rounded-xl p-5 space-y-4 mb-6">
        <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
          <InformationCircleIcon className="w-4 h-4 text-primary" />
          Properties
        </h3>
        <dl className="space-y-3">
          <div>
            <dt className={LABEL}>UUID</dt>
            <dd className="flex items-center gap-1 mt-0.5">
              <span
                className="text-xs font-mono text-text-primary truncate min-w-0"
                title={announcement.uuid}
              >
                {announcement.uuid}
              </span>
              <CopyButton text={announcement.uuid} />
            </dd>
          </div>
          <div>
            <dt className={LABEL}>Date</dt>
            <dd className="text-sm text-text-primary font-medium mt-0.5">
              {formatDateFull(announcement.date)}
            </dd>
          </div>
        </dl>
      </div>

      {/* Content Card */}
      <div className="bg-card border border-border rounded-xl p-5">
        <h3 className="text-xs font-semibold text-text-primary mb-4">
          Content
        </h3>
        <AnnouncementContent
          key={announcement.uuid}
          content={announcement.content}
        />
      </div>

      <DeleteAnnouncementDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        announcement={announcement}
        onDeleted={() => void navigate("/admin/announcements")}
      />
    </div>
  );
}
