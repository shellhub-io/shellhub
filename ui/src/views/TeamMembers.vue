<template>
  <PageHeader
    icon="mdi-account-multiple"
    title="Members"
    overline="Team Management"
    description="Manage team members and their access to this namespace. Invite collaborators and control their permissions."
    icon-color="primary"
    data-test="title"
  >
    <template #actions>
      <MemberInvite @update="getNamespace" />
    </template>
  </PageHeader>

  <div
    class="mt-2"
    data-test="member-list"
  >
    <MemberList />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import axios, { AxiosError } from "axios";
import handleError from "@/utils/handleError";
import MemberInvite from "@/components/Team/Member/MemberInvite.vue";
import MemberList from "@/components/Team/Member/MemberList.vue";
import PageHeader from "@/components/PageHeader.vue";
import useSnackbar from "@/helpers/snackbar";
import useNamespacesStore from "@/store/modules/namespaces";

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const tenant = computed(() => localStorage.getItem("tenant") as string);

const getNamespace = async () => {
  try {
    await namespacesStore.fetchNamespace(tenant.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      const axiosError = error as AxiosError;
      if (axiosError.response?.status === 403) {
        snackbar.showError("You don't have permission to access this resource.");
      }
    } else {
      snackbar.showError("Failed to load namespaces.");
      handleError(error);
    }
  }
};

onMounted(async () => {
  if (tenant.value) {
    await getNamespace();
  }
});
</script>
