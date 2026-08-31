import { z } from "zod";

/**
 * Longest announcement title. It has to fit the modal header without wrapping, which is a
 * narrower limit than the server's.
 */
export const ANNOUNCEMENT_TITLE_MAX = 90;

/**
 * Validates the announcement form. Both fields are checked after trimming, so whitespace alone
 * does not pass as content.
 */
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

/**
 * The announcement form's values, derived from the schema so the two cannot drift.
 */
export type AnnouncementFormValues = z.infer<typeof announcementSchema>;

/**
 * Turns validated form values into the request body, trimming both fields — the stored title
 * must be what was validated, not what was typed around it.
 */
export function buildAnnouncementBody(values: AnnouncementFormValues) {
  return { title: values.title.trim(), content: values.content.trim() };
}
