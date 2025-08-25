import { defineStore } from "pinia";
import { ref } from "vue";
import * as apiAnnouncement from "../api/announcement";
import { IAnnouncement, IAnnouncementShort } from "@/interfaces/IAnnouncement";

const useAnnouncementStore = defineStore("announcement", () => {
  const announcements = ref<IAnnouncementShort[]>([]);
  const currentAnnouncement = ref<IAnnouncement>({} as IAnnouncement);

  const fetchAll = async ({ page, perPage, orderBy }: { page: number, perPage: number, orderBy: "asc" | "desc" }) => {
    try {
      const res = await apiAnnouncement.getListAnnouncements(page, perPage, orderBy);
      announcements.value = res.data as IAnnouncementShort[] ?? [];
      return res.data;
    } catch (error) {
      announcements.value = [];
      throw error;
    }
  };

  const fetchById = async (uuid: string) => {
    try {
      const res = await apiAnnouncement.getAnnouncement(uuid);
      currentAnnouncement.value = res.data as IAnnouncement;
    } catch (error) {
      currentAnnouncement.value = {} as IAnnouncement;
      throw error;
    }
  };

  return {
    announcements,
    currentAnnouncement,
    fetchAll,
    fetchById,
  };
});

export default useAnnouncementStore;
