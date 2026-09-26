import { z } from "zod";

/**
 * The longest login banner, in characters. It must match the server's limit on
 * connection_announcement, which refuses anything longer.
 */
export const BANNER_MAX_LENGTH = 4096;

/**
 * Validates the banner dialog. An empty banner is valid and removes the one set; one past
 * BANNER_MAX_LENGTH is refused with a message the dialog shows beside its counter.
 */
export const bannerSchema = z.object({
  banner: z
    .string()
    .max(
      BANNER_MAX_LENGTH,
      `A banner can be at most ${BANNER_MAX_LENGTH.toLocaleString()} characters.`,
    ),
});

/**
 * The banner dialog's values, inferred from bannerSchema so the two cannot drift apart.
 */
export type BannerFormValues = z.infer<typeof bannerSchema>;
