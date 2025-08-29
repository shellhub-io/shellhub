import { defineStore } from "pinia";
import { ref } from "vue";
import * as announcementApi from "../api/announcement";
import { IAnnouncement, IAnnouncementShort } from "@/interfaces/IAnnouncement";

const useAnnouncementStore = defineStore("announcement", () => {
  const currentAnnouncement = ref<IAnnouncement>({} as IAnnouncement);

  const fetchAnnouncements = async (data?: { page: number, perPage: number, sortOrder?: "asc" | "desc" }) => {
    const res = await announcementApi.fetchAnnouncements(data?.page || 1, data?.perPage || 1, data?.sortOrder || "desc");
    return (res.data ?? []) as IAnnouncementShort[];
  };

  const fetchById = async (uuid: string) => {
    try {
      const res = await announcementApi.getAnnouncement(uuid);
      currentAnnouncement.value = res.data as IAnnouncement;
    } catch (error) {
      currentAnnouncement.value = {} as IAnnouncement;
      throw error;
    }
  };

  return {
    currentAnnouncement,
    fetchAnnouncements,
    fetchById,
  };
});

export default useAnnouncementStore;
