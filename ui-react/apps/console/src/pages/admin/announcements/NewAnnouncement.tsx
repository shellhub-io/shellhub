import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  ChevronRightIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { useAdminCreateAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import AnnouncementEditor from "./AnnouncementEditor";
import { LABEL, INPUT } from "@/utils/styles";

const TITLE_MAX = 90;

export default function NewAnnouncement() {
  const navigate = useNavigate();
  const createAnnouncement = useAdminCreateAnnouncement();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [error, setError] = useState("");

  const titleTrimmed = title.trim();
  const contentTrimmed = content.trim();
  const canSubmit =
    titleTrimmed.length > 0 &&
    titleTrimmed.length <= TITLE_MAX &&
    contentTrimmed.length > 0;

  const handleSubmit = async () => {
    if (!canSubmit) return;
    setError("");
    try {
      await createAnnouncement.mutateAsync({
        body: { title: titleTrimmed, content: contentTrimmed },
      });
      void navigate("/admin/announcements");
    } catch {
      setError("Failed to create announcement. Please try again.");
    }
  };

  return (
    <div>
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/announcements"
          className="text-2xs font-mono text-text-muted hover:text-primary"
        >
          Announcements
        </Link>
        <ChevronRightIcon className="w-3 h-3 text-text-muted/40" />
        <span className="text-2xs font-mono text-text-secondary">New</span>
      </nav>

      {/* Header */}
      <h1 className="text-xl font-semibold text-text-primary mb-6">
        Create Announcement
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
        {/* Title field */}
        <div>
          <label htmlFor="announcement-title" className={LABEL}>
            Title
          </label>
          <input
            id="announcement-title"
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Announcement title"
            maxLength={TITLE_MAX}
            className={INPUT}
            autoFocus
          />
          <div className="flex justify-end mt-1">
            <span className="text-2xs font-mono text-text-muted">
              {titleTrimmed.length}/{TITLE_MAX}
            </span>
          </div>
        </div>

        {/* Content editor */}
        <div>
          <label htmlFor="announcement-content-editor" className={LABEL}>
            Content
          </label>
          <AnnouncementEditor content="" onChange={setContent} />
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-3 pt-2">
          <Link
            to="/admin/announcements"
            className="px-4 py-2.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={!canSubmit || createAnnouncement.isPending}
            className="flex items-center gap-2 px-4 py-2.5 bg-primary text-white text-sm font-medium rounded-lg hover:bg-primary/90 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {createAnnouncement.isPending && (
              <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            )}
            Create
          </button>
        </div>
      </form>
    </div>
  );
}
