<template>
  <div class="d-flex pa-0 align-center" data-test="title">
    <h1>Members</h1>

    <v-spacer />

    <div class="d-flex" data-test="member-invite">
      <MemberInvite @update="refresh" />
    </div>
  </div>

  <div class="mt-2" data-test="member-list">
    <MemberList />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import axios, { AxiosError } from "axios";
import { useStore } from "@/store";
import { INotificationsError } from "@/interfaces/INotifications";
import handleError from "@/utils/handleError";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import MemberList from "@/components/Team/Member/MemberList.vue";

const store = useStore();
const tenant = computed(() => localStorage.getItem("tenant"));

const getNamespace = async () => {
  try {
    await store.dispatch("namespaces/get", tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        store.dispatch("snackbar/showSnackbarErrorAssociation");
      }
    } else {
      store.dispatch(
        "snackbar/showSnackbarErrorAction",
        INotificationsError.namespaceLoad,
      );
      handleError(error);
    }
  }
};

const refresh = () => {
  getNamespace();
};

onMounted(async () => {
  if (tenant.value) {
    await getNamespace();
  }
});
</script>
