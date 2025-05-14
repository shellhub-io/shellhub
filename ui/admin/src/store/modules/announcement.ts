// stores/announcement.ts
import { defineStore } from "pinia";
import { Announcement } from "@admin/api/client/api";
import {
  postAnnouncement,
  updateAnnouncement,
  deleteAnnouncement,
  getAnnouncement,
  getListAnnouncements,
} from "../api/announcement";

interface AnnouncementState {
  announcements: Array<Announcement>;
  announcement: Announcement;
  numberAnnouncements: number;
  page: number;
  perPage: number;
  orderBy: "asc" | "desc";
}

export const useAnnouncementStore = defineStore("announcement", {
  state: (): AnnouncementState => ({
    announcements: [],
    announcement: {} as Announcement,
    numberAnnouncements: 0,
    page: 1,
    perPage: 10,
    orderBy: "asc",
  }),

  getters: {
    getAnnouncements: (state) => state.announcements,
    getAnnouncement: (state) => state.announcement,
    getNumberAnnouncements: (state) => state.numberAnnouncements,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getOrderBy: (state) => state.orderBy,
  },

  actions: {
    async postAnnouncement(announcement: Announcement) {
      try {
        const { data } = await postAnnouncement(announcement as Required<Announcement>);
        this.announcement = data;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async updateAnnouncement(uuid: string, announcement: Announcement) {
      try {
        const { data } = await updateAnnouncement(uuid, announcement as Required<Announcement>);
        this.announcement = data;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async fetchAnnouncement(uuid: string) {
      try {
        const { data } = await getAnnouncement(uuid);
        this.announcement = data;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async fetchAnnouncements({ page, perPage, orderBy }: { page: number; perPage: number; orderBy: "asc" | "desc" }) {
      try {
        const res = await getListAnnouncements(page, perPage, orderBy);
        if (res.data && res.data.length) {
          this.announcements = res.data;
          this.numberAnnouncements = parseInt(res.headers["x-total-count"], 10);
          return res;
        }
        return false;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async deleteAnnouncement(uuid: string) {
      try {
        const { data } = await deleteAnnouncement(uuid);
        this.announcement = data;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    setPageAndPerPage({ page, perPage }: { page: number; perPage: number }) {
      this.page = page;
      this.perPage = perPage;
    },

    setOrderBy(orderBy: "asc" | "desc") {
      this.orderBy = orderBy;
    },

    clearAnnouncements() {
      this.announcements = [];
    },

    clearAnnouncement() {
      this.announcement = {} as Required<Announcement>;
    },
  },
});

export default useAnnouncementStore;
