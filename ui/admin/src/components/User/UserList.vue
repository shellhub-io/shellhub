<template>
  <DataTable
    v-model:items-per-page="itemsPerPage"
    v-model:page="page"
    :headers
    :items="users"
    :loading
    :items-per-page-options="[10, 20, 50, 100]"
    :total-count="userCount"
    table-name="adminUsers"
    data-test="users-list"
  >
    <template #rows>
      <tr
        v-for="(item, i) in users"
        :key="i"
      >
        <td :name-test="item.name">
          {{ item.name }}
        </td>
        <td :email-test="item.email">
          {{ item.email }}
        </td>
        <td :username-test="item.username">
          {{ item.username }}
        </td>
        <td>
          <UserStatusChip :status="item.status" />
        </td>

        <td>
          <v-tooltip
            bottom
            anchor="bottom"
          >
            <template #activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                tabindex="0"
                icon="mdi-information"
                @click="redirectToUser(item)"
                @keyup.enter="redirectToUser(item)"
              />
            </template>
            <span>Info</span>
          </v-tooltip>

          <UserFormDialog
            :key="item.id"
            :user="item"
          />

          <v-tooltip
            bottom
            anchor="bottom"
          >
            <template #activator="{ props }">
              <v-icon
                tag="a"
                dark
                v-bind="props"
                tabindex="0"
                icon="mdi-login"
                @click="loginWithToken(item.id)"
                @keyup.enter="loginWithToken(item.id)"
              />
            </template>
            <span>Login</span>
          </v-tooltip>

          <UserResetPassword
            v-if="userPrefersSAMLAuthentication(item.preferences?.auth_methods)"
            :user-id="item.id"
            @update="fetchUsers"
          />

          <UserDelete
            :id="item.id"
            v-slot="{ openDialog }"
            :show-tooltip="true"
          >
            <v-icon
              icon="mdi-delete"
              tag="button"
              @click="openDialog"
            />
          </UserDelete>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import { IAdminUser, UserAuthMethods } from "@admin/interfaces/IUser";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/Tables/DataTable.vue";
import UserStatusChip from "./UserStatusChip.vue";
import UserFormDialog from "./UserFormDialog.vue";
import UserDelete from "./UserDelete.vue";
import UserResetPassword from "./UserResetPassword.vue";
import handleError from "@/utils/handleError";

const router = useRouter();
const snackbar = useSnackbar();
const usersStore = useUsersStore();
const authStore = useAuthStore();
const page = ref(1);
const itemsPerPage = ref(10);
const loading = ref(false);
const users = computed(() => usersStore.users as IAdminUser[]);
const userCount = computed(() => usersStore.usersCount);
const headers = [
  {
    text: "Name",
    value: "name",
  },
  {
    text: "Email",
    value: "email",
  },
  {
    text: "Username",
    value: "username",
  },
  {
    text: "Status",
    value: "status",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const userPrefersSAMLAuthentication = (authMethods?: UserAuthMethods): boolean =>
  Array.isArray(authMethods) && authMethods.length === 1 && authMethods[0] === "saml";

const fetchUsers = async () => {
  try {
    loading.value = true;
    await usersStore.fetchUsersList({
      perPage: itemsPerPage.value,
      page: page.value,
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch users.");
  } finally {
    loading.value = false;
  }
};

const loginWithToken = async (userId: string) => {
  try {
    const token = await authStore.getLoginToken(userId);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to get the login token.");
  }
};

const redirectToUser = async (user: IAdminUser) => {
  await router.push({ name: "userDetails", params: { id: user.id } });
};

watch([itemsPerPage, page], async () => {
  await fetchUsers();
});

onMounted(async () => {
  await fetchUsers();
});

defineExpose({ users });
</script>
