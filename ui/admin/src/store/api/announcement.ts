import { adminApi } from "@admin/api/http";
import { IAdminAnnouncementRequestBody } from "@admin/interfaces/IAnnouncement";

export const createAnnouncement = async (
  announcement: IAdminAnnouncementRequestBody,
) => adminApi.createAnnouncement(announcement);

export const updateAnnouncement = async (
  uuid: string,
  announcement: IAdminAnnouncementRequestBody,
) => adminApi.updateAnnouncement(uuid, announcement);

export const deleteAnnouncement = async (uuid: string) => adminApi.deleteAnnouncement(uuid);

export const fetchAnnouncementList = async (
  page: number,
  perPage: number,
  orderBy: "asc" | "desc",
) => adminApi.listAnnouncementsAdmin(page, perPage, orderBy);

export const getAnnouncement = async (uuid: string) => adminApi.getAnnouncementAdmin(uuid);
