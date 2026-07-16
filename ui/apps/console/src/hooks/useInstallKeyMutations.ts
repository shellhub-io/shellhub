import { useMutation } from "@tanstack/react-query";
import {
  installKeyCreateMutation,
  installKeyUpdateMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useCreateInstallKey() {
  const invalidate = useInvalidateByIds("installKeyList");
  return useMutation({
    ...installKeyCreateMutation(),
    onSuccess: invalidate,
  });
}

export function useUpdateInstallKey() {
  const invalidate = useInvalidateByIds("installKeyList");
  return useMutation({
    ...installKeyUpdateMutation(),
    onSuccess: invalidate,
  });
}
