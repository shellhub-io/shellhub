import { z } from "zod";

export const ANNOUNCEMENT_TITLE_MAX = 90;

export const announcementSchema = z.object({
  title: z
    .string()
    .refine((t) => t.trim().length > 0, { message: "Title is required" })
    .refine((t) => t.trim().length <= ANNOUNCEMENT_TITLE_MAX, {
      message: `Title must be at most ${ANNOUNCEMENT_TITLE_MAX} characters`,
    }),
  content: z.string().refine((c) => c.trim().length > 0, {
    message: "Content is required",
  }),
});

export type AnnouncementFormValues = z.infer<typeof announcementSchema>;

export function buildAnnouncementBody(values: AnnouncementFormValues) {
  return { title: values.title.trim(), content: values.content.trim() };
}
