<template>
  <div class="d-flex pa-0 align-center">
    <h1>User Details</h1>
    <v-spacer />
    <v-tooltip bottom anchor="bottom">
      <template v-slot:activator="{ props }">
        <v-icon tag="a" v-bind="props" @click="loginWithToken" icon="mdi-login" />
      </template>
      <span>Login</span>
    </v-tooltip>

    <UserDelete :id="userId" redirect />
  </div>

  <v-card class="mt-2 pa-4 bg-background border">
    <v-card-text v-if="currentUser">
      <div v-if="currentUser.status">
        <h3 class="text-overline">Status:</h3>
        <UserStatusChip :status="currentUser.status" />
      </div>
      <div>
        <h3 class="text-overline mt-3">UID:</h3>
        <p :data-test="currentUser.id">{{ currentUser.id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Email:</h3>
        <p :data-test="currentUser.email">{{ currentUser.email }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Username:</h3>
        <p :data-test="currentUser.username">{{ currentUser.username }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Namespaces:</h3>
        <p :data-test="currentUser.namespaces">{{ currentUser.namespaces }}</p>
      </div>
    </v-card-text>
    <p v-else class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onBeforeMount, ref } from "vue";
import { useRoute } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import { IAdminUser } from "@admin/interfaces/IUser";
import useAuthStore from "@admin/store/modules/auth";
import UserStatusChip from "@admin/components/User/UserStatusChip.vue";
import useSnackbar from "@/helpers/snackbar";
import UserDelete from "../components/User/UserDelete.vue";

const route = useRoute();
const snackbar = useSnackbar();
const usersStore = useUsersStore();
const authStore = useAuthStore();
const userId = computed(() => route.params.id as string);
const currentUser = ref({} as IAdminUser);

const loginWithToken = async () => {
  try {
    const token = await authStore.getLoginToken(userId.value);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch { snackbar.showError("Failed to get the login token."); }
};

onBeforeMount(async () => {
  try {
    currentUser.value = await usersStore.fetchUserById(userId.value);
  } catch { snackbar.showError("Failed to get user details."); }
});

defineExpose({ currentUser });
</script>
