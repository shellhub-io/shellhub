import { defineStore } from "pinia";
import { ref } from "vue";
import { IAdminAnnouncement, IAdminAnnouncementRequestBody, IAdminAnnouncementShort } from "@admin/interfaces/IAnnouncement";
import * as announcementApi from "../api/announcement";

const useAnnouncementStore = defineStore("adminAnnouncement", () => {
  const announcements = ref<Array<IAdminAnnouncementShort>>([]);
  const announcement = ref<IAdminAnnouncement>({} as IAdminAnnouncement);
  const announcementCount = ref<number>(0);

  const createAnnouncement = async (announcementData: IAdminAnnouncementRequestBody) => {
    const { data } = await announcementApi.createAnnouncement(announcementData);
    announcement.value = data as IAdminAnnouncement;
  };

  const updateAnnouncement = async (uuid: string, announcementData: IAdminAnnouncementRequestBody) => {
    const { data } = await announcementApi.updateAnnouncement(uuid, announcementData);
    announcement.value = data as IAdminAnnouncement;
  };

  const fetchAnnouncement = async (uuid: string) => {
    const { data } = await announcementApi.getAnnouncement(uuid);
    announcement.value = data as IAdminAnnouncement;
  };

  const fetchAnnouncementList = async (data: { page: number; perPage: number; orderBy: "asc" | "desc"; }) => {
    const res = await announcementApi.fetchAnnouncementList(data.page, data.perPage, data.orderBy);
    announcements.value = res.data as IAdminAnnouncementShort[];
    announcementCount.value = parseInt(res.headers["x-total-count"] as string, 10);
  };

  const deleteAnnouncement = async (uuid: string) => {
    const { data } = await announcementApi.deleteAnnouncement(uuid);
    announcement.value = data as IAdminAnnouncement;
  };

  return {
    announcements,
    announcement,
    announcementCount,
    createAnnouncement,
    updateAnnouncement,
    fetchAnnouncement,
    fetchAnnouncementList,
    deleteAnnouncement,
  };
});

export default useAnnouncementStore;
