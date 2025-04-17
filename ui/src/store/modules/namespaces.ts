import { Module } from "vuex";
import * as apiNamespace from "../api/namespaces";
import { INamespace, INamespaceMember } from "@/interfaces/INamespace";
import { IBilling } from "@/interfaces/IBilling";
import { State } from "..";

export interface NamespacesState {
  namespace: INamespace;
  billing: IBilling;
  namespaces: Array<INamespace>;
  defaultPerPage: number;
  invoicesLength: number;
  numberNamespaces: number;
  owner: boolean;
  userStatus: string;
  invitationLink: string;
}

export const namespaces: Module<NamespacesState, State> = {
  namespaced: true,
  state: {
    namespace: {} as INamespace,
    billing: {} as IBilling,
    namespaces: [],
    defaultPerPage: 3,
    invoicesLength: 0,
    numberNamespaces: 0,
    owner: false,
    userStatus: "",
    invitationLink: "",
  },

  getters: {
    list: (state) => state.namespaces,
    get: (state) => state.namespace,
    getNumberNamespaces: (state) => state.numberNamespaces,
    owner: (state) => state.owner,
    billing: (state) => state.billing,
    getUserStatus: (state) => state.userStatus,
    getInvitationLink: (state) => state.invitationLink,
  },

  mutations: {
    setNamespaces: (state, res) => {
      state.namespaces = res.data;
      state.numberNamespaces = parseInt(res.headers["x-total-count"], 10);
    },

    setNamespace: (state, res) => {
      state.namespace = res.data;
    },

    setBilling: (state, data) => {
      state.billing = data;
    },

    setUserStatus: (state, status) => {
      state.userStatus = status;
    },

    removeNamespace: (state, id) => {
      state.namespaces.splice(
        state.namespaces.findIndex((d) => d.tenant_id === id),
        1,
      );
    },

    removeMember: (state, usr) => {
      state.namespace.members.splice(
        state.namespace.members.findIndex((m: INamespaceMember) => m.username === usr),
        1,
      );
    },

    clearNamespaceList: (state) => {
      state.namespaces = [];
      state.numberNamespaces = 0;
    },

    clearObjectNamespace: (state) => {
      state.namespace = {} as INamespace;
    },

    setOwnerStatus: (state, status) => {
      state.owner = status;
    },

    setInvitationLink: (state, link) => {
      state.invitationLink = link;
    },
  },

  actions: {
    post: async (context, data) => {
      const res = await apiNamespace.postNamespace(data);
      return res;
    },

    fetch: async (context, data) => {
      const res = await apiNamespace.fetchNamespaces(data.page, data.perPage, data.filter);
      context.commit("setNamespaces", res);
    },

    get: async (context, id) => {
      const res = await apiNamespace.getNamespace(id);
      context.commit("setNamespace", res);

      const { billing } = res.data;
      if (billing !== null) {
        context.commit("setBilling", billing);
      }
    },

    put: async (context, data) => {
      const res = await apiNamespace.putNamespace(data);
      context.commit("setNamespace", res);
    },

    remove: async (context, id) => {
      await apiNamespace.removeNamespace(id);
      context.commit("removeNamespace", id);
      context.commit("clearObjectNamespace");
      context.commit("clearNamespaceList");
    },

    leave: async (context, tenant) => {
      const res = await apiNamespace.leaveNamespace(tenant);

      localStorage.setItem("token", res.data.token || "");

      if (res.data.tenant) {
        localStorage.setItem("tenant", res.data.tenant || "");
        localStorage.setItem("role", res.data.role || "");
      }
    },

    sendEmailInvitation: async (context, data) => {
      await apiNamespace.sendNamespaceLink(data);
    },

    generateInvitationLink: async (context, data) => {
      const res = await apiNamespace.generateNamespaceLink(data);
      context.commit("setInvitationLink", res.data.link);
    },

    editUser: async (context, data) => {
      await apiNamespace.editUserToNamespace(data);
    },

    removeUser: async (context, data) => {
      await apiNamespace.removeUserFromNamespace(data);
    },

    acceptInvite: async (context, data) => {
      await apiNamespace.acceptNamespaceInvite(data);
    },

    lookupUserStatus: async (context, data) => {
      const res = await apiNamespace.lookupUserStatus(data);
      context.commit("setUserStatus", res.data?.status);
    },

    clearNamespaceList: (context) => {
      context.commit("clearNamespaceList");
    },

    switchNamespace: async (context, data) => {
      localStorage.removeItem("role");

      const res = await apiNamespace.tenantSwitch(data);
      if (res.status === 200) {
        localStorage.setItem("token", res.data.token || "");
        localStorage.setItem("tenant", data.tenant_id);
        localStorage.setItem("role", res.data.role || "");
      }
    },

    setOwnerStatus: async (context, status) => {
      context.commit("setOwnerStatus", status);
    },
  },
};
