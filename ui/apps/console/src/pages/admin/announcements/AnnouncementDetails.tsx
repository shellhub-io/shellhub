import { useState } from "react";
import { useParams, Link, useNavigate } from "react-router-dom";
import {
  MegaphoneIcon,
  InformationCircleIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { useAdminAnnouncement } from "@/hooks/useAdminAnnouncements";
import Breadcrumb from "@/components/common/Breadcrumb";
import CopyButton from "@/components/common/CopyButton";
import DeleteAnnouncementDialog from "./DeleteAnnouncementDialog";
import AnnouncementContent from "./AnnouncementContent";
import { formatDateFull } from "@/utils/date";
import { PageLoader } from "@shellhub/design-system/components";
import { Button, Card, IconButton } from "@shellhub/design-system/primitives";

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
    return <PageLoader label="Loading announcement details" />;
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
      <Breadcrumb
        items={[
          { label: "Announcements", to: "/admin/announcements" },
          { label: announcement.title },
        ]}
      />

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
          <Button
            variant="outline"
            as={Link}
            to={`/admin/announcements/${announcement.uuid}/edit`}
            icon={<PencilSquareIcon className="w-4 h-4" />}
          >
            Edit
          </Button>
          <IconButton
            variant="danger"
            size="lg"
            aria-label="Delete announcement"
            className="border border-border hover:border-accent-red/30"
            onClick={() => setDeleteOpen(true)}
          >
            <TrashIcon className="w-4 h-4" />
          </IconButton>
        </div>
      </div>

      {/* Properties Card */}
      <Card className="p-5 space-y-4 mb-6">
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
      </Card>

      {/* Content Card */}
      <Card className="p-5">
        <h3 className="text-xs font-semibold text-text-primary mb-4">
          Content
        </h3>
        <AnnouncementContent
          key={announcement.uuid}
          content={announcement.content}
        />
      </Card>

      <DeleteAnnouncementDialog
        open={deleteOpen}
        onClose={() => setDeleteOpen(false)}
        announcement={announcement}
        onDeleted={() => void navigate("/admin/announcements")}
      />
    </div>
  );
}
