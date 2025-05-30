<template>
  <div>
    <DataTable
      :headers="header"
      :items="users"
      :itemsPerPage="itemsPerPage"
      :nextPage="next"
      :previousPage="prev"
      :loading="loading"
      :comboboxOptions="[10, 20, 50, 100]"
      :totalCount="totalUsers"
      :actualPage="page"
      @changeItemsPerPage="changeItemsPerPage"
      @clickNextPage="next"
      @clickPreviousPage="prev"
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
          <td v-if="item.confirmed" class="pl-0">
            <v-chip class="ma-2" color="success" variant="text" prepend-icon="mdi-checkbox-marked-circle">
              Confirmed
            </v-chip>
          </td>
          <td v-else class="pl-0">
            <v-chip class="ma-2" color="warning" variant="text" prepend-icon="mdi-alert-circle">
              Not confirmed
            </v-chip>
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
                  @click="loginToken(item)"
                  tabindex="0"
                  @keyup.enter="loginToken(item)"
                >mdi-login
                </v-icon>
              </template>
              <span>Login</span>
            </v-tooltip>

            <UserResetPassword
              v-if="checkAuthMethods(item as IUser)"
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
import { UserAdminResponse } from "@admin/api/client/api";
import useAuthStore from "@admin/store/modules/auth";
import useSnackbar from "@/helpers/snackbar";
import DataTable from "../DataTable.vue";
import UserFormDialog from "./UserFormDialog.vue";
import UserDelete from "./UserDelete.vue";
import UserResetPassword from "./UserResetPassword.vue";

export interface IUser {
  id: string;
  auth_methods: Array<string>;
  namespaces: number;
  confirmed: boolean;
  created_at: string;
  last_login: string;
  name: string;
  email: string;
  username: string;
  password: string;
}

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

const header = [
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

const checkAuthMethods = (user: IUser | undefined) => user?.auth_methods
  && user.auth_methods.length === 1
  && user.auth_methods[0] === "saml";

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
    const hasUsers = await userStore.fetch({
      perPage: perPageValue,
      page: pageValue,
      filter: filter.value,
    });

    if (!hasUsers) {
      page.value--;
    }

    loading.value = false;
  } catch (error) {
    snackbar.showError("Failed to fetch users.");
  }
};

const next = async () => {
  await getUsers(itemsPerPage.value, ++page.value);
};

const prev = async () => {
  if (page.value > 1) await getUsers(itemsPerPage.value, --page.value);
};

const changeItemsPerPage = async (newItemsPerPage: number) => {
  itemsPerPage.value = newItemsPerPage;
};

watch(itemsPerPage, () => {
  getUsers(itemsPerPage.value, page.value);
});

const loginToken = async (user) => {
  try {
    const token = await authStore.loginToken(user);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch {
    snackbar.showError("Failed to get the login token.");
  }
};

const refreshUsers = async () => {
  await userStore.refresh();
};

const redirectToUser = async (user: UserAdminResponse) => {
  router.push({ name: "userDetails", params: { id: user.id } });
};

defineExpose({ users });
</script>
