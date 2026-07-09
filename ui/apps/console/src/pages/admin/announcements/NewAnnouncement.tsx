import { Link, useNavigate } from "react-router-dom";
import { useForm, useController, useWatch } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAdminCreateAnnouncement } from "@/hooks/useAdminAnnouncementMutations";
import AnnouncementEditor from "./AnnouncementEditor";
import Breadcrumb from "@/components/common/Breadcrumb";
import { FormInputField } from "@/components/common/fields/rhf";
import FieldLabel from "@/components/common/fields/FieldLabel";
import { Button, Callout, Card } from "@shellhub/design-system/primitives";
import {
  announcementSchema,
  buildAnnouncementBody,
  ANNOUNCEMENT_TITLE_MAX,
  type AnnouncementFormValues,
} from "./announcementSchema";

export default function NewAnnouncement() {
  const navigate = useNavigate();
  const createAnnouncement = useAdminCreateAnnouncement();

  const form = useForm<AnnouncementFormValues>({
    mode: "onChange",
    resolver: zodResolver(announcementSchema),
    defaultValues: { title: "", content: "" },
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

  const onValid = async (values: AnnouncementFormValues) => {
    clearErrors("root");
    try {
      await createAnnouncement.mutateAsync({
        body: buildAnnouncementBody(values),
      });
      void navigate("/admin/announcements");
    } catch {
      setError("root", {
        message: "Failed to create announcement. Please try again.",
      });
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

      <h1 className="text-xl font-semibold text-text-primary mb-6">
        Create Announcement
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
          <AnnouncementEditor content="" onChange={contentField.onChange} />
        </div>

        <div className="flex items-center justify-end gap-3 pt-2">
          <Link
            to="/admin/announcements"
            className="px-4 py-2.5 text-sm text-text-secondary hover:text-text-primary transition-colors"
          >
            Cancel
          </Link>
          <Button
            type="submit"
            loading={createAnnouncement.isPending}
            disabled={!isValid || createAnnouncement.isPending}
          >
            Create
          </Button>
        </div>
      </Card>
    </div>
  );
}
