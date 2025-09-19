<template>
  <div class="d-flex flex-column justify-space-between align-center flex-sm-row">
    <h1 class="mr-2">Users</h1>
    <div class="w-50">
      <v-text-field
        label="Search by username"
        variant="underlined"
        color="primary"
        single-line
        hide-details
        v-model.trim="filter"
        v-on:keyup.enter="searchUsers"
        append-inner-icon="mdi-magnify"
        @click:append-inner="searchUsers"
        density="comfortable"
      />
    </div>
    <div class="d-flex mt-4">
      <UserExport class="ml-2" data-test="users-export-btn" />
      <UserFormDialog title-card="Create new user" create-user />
    </div>
  </div>
  <v-card class="mt-2">
    <UserList />
  </v-card>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { watchDebounced } from "@vueuse/core";
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

watchDebounced(filter, () => {
  searchUsers();
}, { debounce: 1000, maxWait: 5000 });

defineExpose({ filter });
</script>
