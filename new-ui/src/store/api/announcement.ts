import { announcementApi } from "@/api/http";

export const getListAnnouncements = async (
  page : number,
  perPage: number,
  orderBy: "asc" | "desc",
) => {
  return announcementApi.listAnnouncements(page, perPage, orderBy);
};

export const getAnnouncement = async (uuid: string) => announcementApi.getAnnouncement(uuid);
