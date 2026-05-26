import { useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  ExclamationCircleIcon,
  MegaphoneIcon,
} from "@heroicons/react/24/outline";
import { useAdminAnnouncement } from "@/hooks/useAdminAnnouncements";
import { useAdminUpdateAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import AnnouncementEditor from "./AnnouncementEditor";
import Breadcrumb from "@/components/common/Breadcrumb";
import InputField from "@/components/common/fields/InputField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import Spinner from "@/components/common/Spinner";
import PageLoader from "@/components/common/PageLoader";

const TITLE_MAX = 90;

export default function EditAnnouncement() {
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const {
    data: announcement,
    isLoading: isFetching,
    error: fetchError,
  } = useAdminAnnouncement(uuid ?? "");
  const updateAnnouncement = useAdminUpdateAnnouncement();

  const [title, setTitle] = useState<string | null>(null);
  const [content, setContent] = useState<string | null>(null);
  const [error, setError] = useState("");

  // Use fetched data as defaults until user edits
  const currentTitle = title ?? announcement?.title ?? "";
  const currentContent = content ?? announcement?.content ?? "";
  const titleTrimmed = currentTitle.trim();
  const contentTrimmed = currentContent.trim();

  const canSubmit =
    titleTrimmed.length > 0 &&
    titleTrimmed.length <= TITLE_MAX &&
    contentTrimmed.length > 0;

  const handleSubmit = async () => {
    if (!canSubmit || !uuid) return;
    setError("");
    try {
      await updateAnnouncement.mutateAsync({
        path: { uuid },
        body: { title: titleTrimmed, content: contentTrimmed },
      });
      void navigate(`/admin/announcements/${uuid}`);
    } catch {
      setError("Failed to update announcement. Please try again.");
    }
  };

  if (isFetching) {
    return (
      <PageLoader label="Loading announcement" />
    );
  }

  if (fetchError || !announcement) {
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
    <div>
      <Breadcrumb
        items={[
          { label: "Announcements", to: "/admin/announcements" },
          { label: announcement.title, to: `/admin/announcements/${uuid}` },
          { label: "Edit" },
        ]}
      />

      {/* Header */}
      <h1 className="text-xl font-semibold text-text-primary mb-6">
        Edit Announcement
      </h1>

      {error && (
        <div
          role="alert"
          className="flex items-center gap-2 bg-accent-red/8 border border-accent-red/20 text-accent-red px-3.5 py-2.5 rounded-md text-xs font-mono mb-4 animate-slide-down"
        >
          <ExclamationCircleIcon
            className="w-3.5 h-3.5 shrink-0"
            strokeWidth={2}
          />
          {error}
        </div>
      )}

      {/* Form */}
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void handleSubmit();
        }}
        className="bg-card border border-border rounded-xl p-6 space-y-5"
      >
        <InputField
          id="announcement-title"
          label="Title"
          labelAdornment={
            <span className="ml-auto text-2xs font-mono text-text-muted">
              {titleTrimmed.length}/{TITLE_MAX}
            </span>
          }
          value={currentTitle}
          onChange={setTitle}
          placeholder="Announcement title"
          maxLength={TITLE_MAX}
          autoFocus
        />

        {/* Content editor */}
        <div>
          <FieldLabel htmlFor="announcement-content-editor">Content</FieldLabel>
          <AnnouncementEditor
            key={announcement.uuid}
            content={announcement.content}
            onChange={setContent}
          />
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Link
            to={`/admin/announcements/${uuid}`}
            className="px-4 py-2.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={!canSubmit || updateAnnouncement.isPending}
            className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {updateAnnouncement.isPending && (
              <Spinner size="sm" tone="onPrimary" />
            )}
            Save
          </button>
        </div>
      </form>
    </div>
  );
}
