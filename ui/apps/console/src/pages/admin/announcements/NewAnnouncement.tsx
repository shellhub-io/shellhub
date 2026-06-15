import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import Alert from "@/components/common/Alert";
import { useAdminCreateAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import AnnouncementEditor from "./AnnouncementEditor";
import Breadcrumb from "@/components/common/Breadcrumb";
import InputField from "@/components/common/fields/InputField";
import FieldLabel from "@/components/common/fields/FieldLabel";
import { Card, Spinner } from "@shellhub/design-system/primitives";

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
      <Breadcrumb
        items={[
          { label: "Announcements", to: "/admin/announcements" },
          { label: "New" },
        ]}
      />

      {/* Header */}
      <h1 className="text-xl font-semibold text-text-primary mb-6">
        Create Announcement
      </h1>

      {error && (
        <Alert variant="error" className="mb-4">
          {error}
        </Alert>
      )}

      {/* Form */}
      <Card
        as="form"
        onSubmit={(e: React.FormEvent) => {
          e.preventDefault();
          void handleSubmit();
        }}
        className="p-6 space-y-5"
      >
        <InputField
          id="announcement-title"
          label="Title"
          labelAdornment={
            <span className="ml-auto text-2xs font-mono text-text-muted">
              {titleTrimmed.length}/{TITLE_MAX}
            </span>
          }
          value={title}
          onChange={setTitle}
          placeholder="Announcement title"
          maxLength={TITLE_MAX}
          autoFocus
        />

        {/* Content editor */}
        <div>
          <FieldLabel htmlFor="announcement-content-editor">Content</FieldLabel>
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
              <Spinner size="sm" tone="onPrimary" />
            )}
            Create
          </button>
        </div>
      </Card>
    </div>
  );
}
