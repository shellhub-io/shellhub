<template>
  <div class="d-flex pa-0 align-center">
    <h1>User Details</h1>

    <v-col class="pr-4 text-right">
      <v-tooltip bottom anchor="bottom">
        <template v-slot:activator="{ props }">
          <v-icon tag="a" dark v-bind="props" @click="loginToken"> mdi-login </v-icon>
        </template>
        <span>Login</span>
      </v-tooltip>

      <UserDelete :id="userId" redirect />
    </v-col>
  </div>

  <v-card v-if="currentUser" class="mt-2 pa-4">
    <v-card-text>
      <div class="text-overline mt-3" v-if="currentUser.status">
        <h3>Status:</h3>
        <UserStatusChip :status="currentUser.status" />
      </div>
      <div>
        <div class="text-overline mt-3">
          <h3>uid:</h3>
        </div>
        <div :data-test="currentUser.id">
          <p>{{ currentUser.id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Email:</h3>
        </div>
        <div :data-test="currentUser.email">
          <p>{{ currentUser.email }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Username:</h3>
        </div>
        <div :data-test="currentUser.username">
          <p>{{ currentUser.username }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Namespaces:</h3>
        </div>
        <div :data-test="currentUser.namespaces">
          <p>{{ currentUser.namespaces }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onBeforeMount } from "vue";
import { useRoute } from "vue-router";
import useUsersStore from "@admin/store/modules/users";
import { IUser } from "@admin/interfaces/IUser";
import useAuthStore from "@admin/store/modules/auth";
import UserStatusChip from "@admin/components/User/UserStatusChip.vue";
import useSnackbar from "@/helpers/snackbar";
import UserDelete from "../components/User/UserDelete.vue";

const route = useRoute();
const snackbar = useSnackbar();
const userStore = useUsersStore();
const authStore = useAuthStore();
const userId = computed(() => route.params.id as string);
const currentUser = computed(() => userStore.getUser as IUser);

const loginToken = async () => {
  try {
    const token = await authStore.loginToken(userId.value);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch {
    snackbar.showError("Failed to get the login token.");
  }
};

onBeforeMount(async () => {
  try {
    await userStore.get(userId.value);
  } catch {
    snackbar.showError("Failed to get user details.");
  }
});

defineExpose({ currentUser });
</script>
