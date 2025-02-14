import { adminApi } from "./../../api/http";

type Announcement = {
  title: string;
  content: string;
};

const postAnnouncement = async (
  announcement: Announcement,
) => adminApi.createAnnouncement(announcement);

const updateAnnouncement = async (
  uuid: string,
  announcement: Announcement,
) => adminApi.updateAnnouncement(uuid, announcement);

const deleteAnnouncement = async (uuid: string) => adminApi.deleteAnnouncement(uuid);

const getListAnnouncements = async (
  page: number,
  perPage: number,
  orderBy: "asc" | "desc",
) => adminApi.listAnnouncementsAdmin(page, perPage, orderBy);

const getAnnouncement = async (uuid: string) => adminApi.getAnnouncementAdmin(uuid);

export {
  postAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  getListAnnouncements,
  getAnnouncement,
};
