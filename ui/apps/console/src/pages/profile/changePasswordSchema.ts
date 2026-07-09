import { z } from "zod";
import { validatePassword } from "@/utils/validation";

/**
 * Only surfaces format errors for fields the user has actually filled — the
 * "all fields required" gate is enforced at the submit button (see the drawer's
 * `submitDisabled`), so a partially-filled form blocks submit without spraying
 * required-field errors across untouched inputs.
 */
export const changePasswordSchema = z
  .object({
    current: z.string(),
    newPw: z.string(),
    confirmPw: z.string(),
  })
  .superRefine((values, ctx) => {
    if (values.newPw) {
      const newPwError = validatePassword(values.newPw);
      if (newPwError) ctx.addIssue({ code: "custom", path: ["newPw"], message: newPwError });
    }

    if (values.confirmPw && values.confirmPw !== values.newPw) {
      ctx.addIssue({ code: "custom", path: ["confirmPw"], message: "Passwords do not match" });
    }
  });

export type ChangePasswordFormValues = z.infer<typeof changePasswordSchema>;
