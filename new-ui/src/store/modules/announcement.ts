import { Module } from "vuex";
import { State } from "./../index";
import * as apiAnnouncement from "../api/announcement";

export interface AnnouncementState {
  announcements: Array<any>;
  announcement: any;
  page: number;
  perPage: number;
  orderBy: "asc" | "desc";
}

export const announcement: Module<AnnouncementState, State> = {
  namespaced: true,
  state: {
    announcements: [],
    announcement: {},
    page: 1,
    perPage: 10,
    orderBy: "asc",
  },
  getters: {
    list: (state) => state.announcements,
    get: (state) => state.announcement,
    getPage: (state) => state.page,
    getPerPage: (state) => state.perPage,
    getOrderBy: (state) => state.orderBy,
  },
  mutations: {
    setAnnouncements: (state, res) => {
      state.announcements = res;
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
      state.announcement = {};
    },
  },

  actions: {
    async getListAnnouncements({ commit }, { page, perPage, orderBy }) {
      try {
        const res = await apiAnnouncement.getListAnnouncements(page, perPage, orderBy);
  
        if (res.data) {
          commit("setAnnouncements", res.data);
          commit("setPageAndPerPage", { page, perPage });
          commit("setOrderBy", orderBy);
          return res.data;
        }
  
        commit("clearAnnouncements");
        return false;
      } catch (error) {
        commit("clearAnnouncements");
        throw error;;
      }
    },

    async getAnnouncement({ commit }, uuid) {
      try {
        const res = await apiAnnouncement.getAnnouncement(uuid);
        commit("setAnnouncement", res.data);
      } catch (error) {
        commit("clearAnnouncement");
        throw error;
      }
    }
  },
}
