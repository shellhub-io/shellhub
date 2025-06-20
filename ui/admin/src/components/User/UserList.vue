<template>
  <div>
    <DataTable
      :headers
      :items="users"
      v-model:itemsPerPage="itemsPerPage"
      v-model:page="page"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      :totalCount="userCount"
      data-test="users-list"
    >
      <template v-slot:rows>
        <tr v-for="(item, i) in users" :key="i">
          <td :name-test="item.name">
            {{ item.name }}
          </td>
          <td :email-test="item.email">
            {{ item.email }}
          </td>
          <td :username-test="item.username">
            {{ item.username }}
          </td>
          <td :namespaces-test="item.namespaces">
            {{ item.namespaces }}
          </td>
          <td>
            <UserStatusChip :status="item.status" />
          </td>

          <td>
            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="redirectToUser(item)"
                  @keyup.enter="redirectToUser(item)"
                  tabindex="0"
                  icon="mdi-information"
                />
              </template>
              <span>Info</span>
            </v-tooltip>

            <UserFormDialog titleCard="Edit user" :user="item" />

            <v-tooltip bottom anchor="bottom">
              <template v-slot:activator="{ props }">
                <v-icon
                  tag="a"
                  dark
                  v-bind="props"
                  @click="loginToken(item.id)"
                  tabindex="0"
                  @keyup.enter="loginToken(item.id)"
                  icon="mdi-login"
                />
              </template>
              <span>Login</span>
            </v-tooltip>

            <UserResetPassword
              v-if="userPrefersSAMLAuthentication(item.preferences.auth_methods)"
              :userId="item.id"
              @update="refreshUsers"
            />

            <UserDelete :id="item.id" @update="refreshUsers" />
          </td>
        </tr>
      </template>
    </DataTable>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import { IUser, UserAuthMethods } from "@admin/interfaces/IUser";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "@/components/DataTable.vue";
import UserStatusChip from "./UserStatusChip.vue";
import UserFormDialog from "./UserFormDialog.vue";
import UserDelete from "./UserDelete.vue";
import UserResetPassword from "./UserResetPassword.vue";
import handleError from "@/utils/handleError";

const router = useRouter();
const snackbar = useSnackbar();
const userStore = useUsersStore();
const authStore = useAuthStore();
const page = ref(1);
const itemsPerPage = ref(10);
const loading = ref(false);
const users = computed(() => userStore.getUsers as unknown as IUser[]);
const userCount = computed(() => userStore.numberUsers);
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
    text: "Namespaces",
    value: "namespaces",
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

const userPrefersSAMLAuthentication = (authMethods: UserAuthMethods) => (
  authMethods && authMethods.length === 1 && authMethods[0] === "saml"
);

const fetchUsers = async () => {
  try {
    loading.value = true;
    await userStore.fetch({
      perPage: itemsPerPage.value,
      page: page.value,
      filter: "",
    });
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to fetch users.");
  }
  loading.value = false;
};

const loginToken = async (userId: string) => {
  try {
    const token = await authStore.loginToken(userId);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to get the login token.");
  }
};

const refreshUsers = async () => {
  await userStore.refresh();
};

const redirectToUser = async (user: IUser) => {
  router.push({ name: "userDetails", params: { id: user.id } });
};

watch([itemsPerPage, page], async () => {
  await fetchUsers();
});

onMounted(async () => {
  await fetchUsers();
});

defineExpose({ users });
</script>
