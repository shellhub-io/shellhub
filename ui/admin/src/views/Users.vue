<template>
  <PageHeader
    icon="mdi-account-group"
    title="Users"
    overline="Account Management"
    description="Review every account, keep credentials healthy, and quickly onboard administrators."
    icon-color="primary"
  >
    <template #actions>
      <div class="d-flex flex-column flex-sm-row ga-2 mt-2 mt-md-0">
        <UserExport data-test="users-export-btn" />
        <UserFormDialog create-user />
      </div>
    </template>

    <v-text-field
      v-model.trim="filter"
      label="Search by username"
      color="primary"
      class="w-100 w-md-50"
      single-line
      hide-details
      append-inner-icon="mdi-magnify"
      density="compact"
      @keyup.enter="searchUsers"
      @click:append-inner="searchUsers"
    />
  </PageHeader>
  <UserList />
</template>

<script setup lang="ts">
import { ref } from "vue";
import { watchDebounced } from "@vueuse/core";
import PageHeader from "@/components/PageHeader.vue";
import useUsersStore from "@admin/store/modules/users";
import UserList from "../components/User/UserList.vue";
import UserFormDialog from "../components/User/UserFormDialog.vue";
import UserExport from "../components/User/UserExport.vue";

const usersStore = useUsersStore();
const filter = ref("");

const searchUsers = async () => {
  const filterToEncodeBase64 = [
    { type: "property", params: { name: "username", operator: "contains", value: filter.value } },
  ];

  const encodedFilter = filter.value ? btoa(JSON.stringify(filterToEncodeBase64)) : "";

  usersStore.setFilter(encodedFilter);

  await usersStore.fetchUsersList({ filter: encodedFilter });
};

watchDebounced(filter, async () => {
  await searchUsers();
}, { debounce: 1000, maxWait: 5000 });

defineExpose({ filter });
</script>
