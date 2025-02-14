import { Module } from "vuex";
import { IAnnouncement, IAnnouncements } from "./../../interfaces/IAnnouncements";
import { State } from "./../index";
import { postAnnouncement, updateAnnouncement, deleteAnnouncement, getAnnouncement, getListAnnouncements } from "../api/announcement";

export interface AnnouncementState {
  announcements: Array<IAnnouncements>;
  announcement: IAnnouncement;
  numberAnnouncements: number;
  page: number;
  perPage: number;
  orderBy: "asc" | "desc";
}

export const announcement: Module<AnnouncementState, State> = {
  namespaced: true,
  state: {
    announcements: [],
    announcement: {} as IAnnouncement,
    numberAnnouncements: 0,
    page: 1,
    perPage: 10,
    orderBy: "asc",
  },
  getters: {
    announcements: (state) => state.announcements,
    announcement: (state) => state.announcement,
    numberAnnouncements: (state) => state.numberAnnouncements,
    page: (state) => state.page,
    perPage: (state) => state.perPage,
    orderBy: (state) => state.orderBy,
  },
  mutations: {
    setAnnouncements: (state, res) => {
      state.announcements = res.data;
      state.numberAnnouncements = parseInt(res.headers["x-total-count"], 10);
    },

    setAnnouncement: (state, res) => {
      state.announcement = res;
    },

    setPageAndPerPage: (state, { page, perPage }) => {
      state.page = page;
      state.perPage = perPage;
    },

    setOrderBy: (state, orderBy) => {
      state.orderBy = orderBy;
    },

    clearAnnouncements: (state) => {
      state.announcements = [];
    },

    clearAnnouncement: (state) => {
      state.announcement = {} as IAnnouncement;
    },
  },
  actions: {
    async postAnnouncement({ commit }, announcement) {
      try {
        const { data } = await postAnnouncement(announcement);
        commit("setAnnouncement", data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async updateAnnouncement({ commit }, { uuid, announcement }) {
      try {
        const { data } = await updateAnnouncement(uuid, announcement);
        commit("setAnnouncement", data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async getAnnouncement({ commit }, uuid) {
      try {
        const { data } = await getAnnouncement(uuid);
        commit("setAnnouncement", data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async getAnnouncements({ commit }, { page, perPage, orderBy }) {
      try {
        const res = await getListAnnouncements(page, perPage, orderBy);

        if (res.data && res.data.length) {
          commit("setAnnouncements", res);
          return res;
        }
        return false;
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

    async deleteAnnouncement({ commit }, uuid) {
      try {
        const { data } = await deleteAnnouncement(uuid);
        commit("setAnnouncement", data);
      } catch (error) {
        console.error(error);
        throw error;
      }
    },

  },
};
