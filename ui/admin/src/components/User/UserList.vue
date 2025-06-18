<template>
  <div>
    <DataTable
      :headers
      :items="users"
      v-model:itemsPerPage="itemsPerPage"
      v-model:page="page"
      :loading
      :itemsPerPageOptions="[10, 20, 50, 100]"
      :totalCount="totalUsers"
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
                >mdi-information
                </v-icon>
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
                >mdi-login
                </v-icon>
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

const router = useRouter();
const snackbar = useSnackbar();
const userStore = useUsersStore();
const authStore = useAuthStore();

const itemsPerPage = ref(10);
const loading = ref(false);
const page = ref(1);
const filter = ref("");
const users = computed(() => userStore.getUsers as unknown as IUser[]);
const totalUsers = computed(() => userStore.numberUsers);

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

onMounted(async () => {
  try {
    loading.value = true;
    await userStore.fetch({
      perPage: itemsPerPage.value,
      page: page.value,
      filter: filter.value,
    });
  } catch (error) {
    snackbar.showError("Failed to fetch users.");
  } finally {
    loading.value = false;
  }
});

const getUsers = async (perPageValue: number, pageValue: number) => {
  try {
    loading.value = true;
    await userStore.fetch({
      perPage: perPageValue,
      page: pageValue,
      filter: filter.value,
    });

    loading.value = false;
  } catch (error) {
    snackbar.showError("Failed to fetch users.");
  }
};

watch([itemsPerPage, page], () => {
  getUsers(itemsPerPage.value, page.value);
});

const loginToken = async (userId: string) => {
  try {
    const token = await authStore.loginToken(userId);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch {
    snackbar.showError("Failed to get the login token.");
  }
};

const refreshUsers = async () => {
  await userStore.refresh();
};

const redirectToUser = async (user: IUser) => {
  router.push({ name: "userDetails", params: { id: user.id } });
};

defineExpose({ users });
</script>
