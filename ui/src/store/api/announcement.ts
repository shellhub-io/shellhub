import { announcementApi } from "@/api/http";

export const fetchAnnouncements = async (
  page: number,
  perPage: number,
  sortOrder?: "asc" | "desc",
) => announcementApi.listAnnouncements(page, perPage, sortOrder);

export const getAnnouncement = async (uuid: string) => announcementApi.getAnnouncement(uuid);
