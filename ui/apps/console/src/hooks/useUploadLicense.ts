import { useMutation } from "@tanstack/react-query";
import { sendLicenseMutation } from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useUploadLicense() {
  const invalidate = useInvalidateByIds("getLicense");
  return useMutation({
    ...sendLicenseMutation(),
    onSuccess: invalidate,
  });
}
