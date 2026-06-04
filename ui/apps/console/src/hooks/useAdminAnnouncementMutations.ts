import { useMutation } from "@tanstack/react-query";
import {
  createAnnouncementMutation,
  updateAnnouncementMutation,
  deleteAnnouncementMutation,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";

export function useAdminCreateAnnouncement() {
  const invalidate = useInvalidateByIds("listAnnouncementsAdmin");
  return useMutation({
    ...createAnnouncementMutation(),
    onSuccess: invalidate,
  });
}

export function useAdminUpdateAnnouncement() {
  const invalidate = useInvalidateByIds(
    "listAnnouncementsAdmin",
    "getAnnouncementAdmin",
  );
  return useMutation({
    ...updateAnnouncementMutation(),
    onSuccess: invalidate,
  });
}

export function useAdminDeleteAnnouncement() {
  const invalidate = useInvalidateByIds(
    "listAnnouncementsAdmin",
    "getAnnouncementAdmin",
  );
  return useMutation({
    ...deleteAnnouncementMutation(),
    onSuccess: invalidate,
  });
}
