import { useMutation } from "@tanstack/react-query";
import {
  createAnnouncementMutation,
  updateAnnouncementMutation,
  deleteAnnouncementMutation,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";

/**
 * Creates an announcement, refreshing the admin list on success.
 */
export function useAdminCreateAnnouncement() {
  const invalidate = useInvalidateByIds("listAnnouncementsAdmin");
  return useMutation({
    ...createAnnouncementMutation(),
    onSuccess: invalidate,
  });
}

/**
 * Updates an announcement, refreshing both the list and the single-announcement query.
 */
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

/**
 * Deletes an announcement, refreshing both the list and the single-announcement query.
 */
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
