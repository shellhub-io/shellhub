import { create } from "zustand";
import {
  getMembers,
  addMember,
  updateMemberRole,
  removeMember,
  type NamespaceMember,
} from "../api/team";

interface MembersState {
  members: NamespaceMember[];
  loading: boolean;
  fetch: (tenantId: string) => Promise<void>;
  addMember: (tenantId: string, email: string, role: string) => Promise<void>;
  updateRole: (tenantId: string, userId: string, role: string) => Promise<void>;
  remove: (tenantId: string, userId: string) => Promise<void>;
}

export const useMembersStore = create<MembersState>((set, get) => ({
  members: [],
  loading: false,

  fetch: async (tenantId) => {
    set({ loading: true });
    try {
      const members = await getMembers(tenantId);
      set({ members, loading: false });
    } catch {
      set({ loading: false });
    }
  },

  addMember: async (tenantId, email, role) => {
    await addMember(tenantId, { email, role });
    await get().fetch(tenantId);
  },

  updateRole: async (tenantId, userId, role) => {
    await updateMemberRole(tenantId, userId, role);
    await get().fetch(tenantId);
  },

  remove: async (tenantId, userId) => {
    await removeMember(tenantId, userId);
    await get().fetch(tenantId);
  },
}));
