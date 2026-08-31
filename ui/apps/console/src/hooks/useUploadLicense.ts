import { useMutation } from "@tanstack/react-query";
import { sendLicenseMutation } from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Uploads a licence file and refreshes the licence query, so the new terms take effect in the UI
 * as soon as the server accepts them.
 */
export function useUploadLicense() {
  const invalidate = useInvalidateByIds("getLicense");
  return useMutation({
    ...sendLicenseMutation(),
    onSuccess: invalidate,
  });
}
