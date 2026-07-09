import { useMemo } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useForm, useController, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { MegaphoneIcon } from "@heroicons/react/24/outline";
import { useAdminAnnouncement } from "@/hooks/useAdminAnnouncements";
import { useAdminUpdateAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import AnnouncementEditor from "./AnnouncementEditor";
import Breadcrumb from "@/components/common/Breadcrumb";
import { FormInputField } from "@/components/common/fields/rhf";
import FieldLabel from "@/components/common/fields/FieldLabel";
import PageLoader from "@/components/common/PageLoader";
import { Button, Callout, Card } from "@shellhub/design-system/primitives";
import {
  announcementSchema,
  buildAnnouncementBody,
  ANNOUNCEMENT_TITLE_MAX,
  type AnnouncementFormValues,
} from "./announcementSchema";

export default function EditAnnouncement() {
  const { uuid } = useParams<{ uuid: string }>();
  const navigate = useNavigate();
  const {
    data: announcement,
    isLoading: isFetching,
    error: fetchError,
  } = useAdminAnnouncement(uuid ?? "");
  const updateAnnouncement = useAdminUpdateAnnouncement();

  const values = useMemo<AnnouncementFormValues>(
    () => ({
      title: announcement?.title ?? "",
      content: announcement?.content ?? "",
    }),
    [announcement],
  );

  const form = useForm<AnnouncementFormValues>({
    mode: "onChange",
    resolver: zodResolver(announcementSchema),
    values,
    resetOptions: { keepDirtyValues: true },
  });
  const {
    control,
    handleSubmit,
    setError,
    clearErrors,
    formState: { isValid, errors },
  } = form;

  const { field: contentField } = useController({ name: "content", control });
  const title = useWatch({ control, name: "title" });
  const titleLength = title.trim().length;

  const onValid = async (formValues: AnnouncementFormValues) => {
    if (!uuid) return;
    clearErrors("root");
    try {
      await updateAnnouncement.mutateAsync({
        path: { uuid },
        body: buildAnnouncementBody(formValues),
      });
      void navigate(`/admin/announcements/${uuid}`);
    } catch {
      setError("root", {
        message: "Failed to update announcement. Please try again.",
      });
    }
  };

  if (isFetching) {
    return <PageLoader label="Loading announcement" />;
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

      <h1 className="text-xl font-semibold text-text-primary mb-6">
        Edit Announcement
      </h1>

      {errors.root && (
        <Callout variant="error" className="mb-4">
          {errors.root.message}
        </Callout>
      )}

      <Card
        as="form"
        onSubmit={(e: React.FormEvent) => void handleSubmit(onValid)(e)}
        className="p-6 space-y-5"
      >
        <FormInputField
          name="title"
          control={control}
          id="announcement-title"
          label="Title"
          labelAdornment={
            <span className="ml-auto text-2xs font-mono text-text-muted">
              {titleLength}/{ANNOUNCEMENT_TITLE_MAX}
            </span>
          }
          placeholder="Announcement title"
          maxLength={ANNOUNCEMENT_TITLE_MAX}
        />

        <div>
          <FieldLabel htmlFor="announcement-content-editor">Content</FieldLabel>
          <AnnouncementEditor
            key={announcement.uuid}
            content={announcement.content}
            onChange={contentField.onChange}
          />
        </div>

        <div className="flex items-center justify-end gap-3 pt-2">
          <Link
            to={`/admin/announcements/${uuid}`}
            className="px-4 py-2.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </Link>
          <Button
            type="submit"
            loading={updateAnnouncement.isPending}
            disabled={!isValid || updateAnnouncement.isPending}
          >
            Save
          </Button>
        </div>
      </Card>
    </div>
  );
}
