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
      <div class="text-overline mt-3">
        <h3>Status:</h3>
        <v-chip
          v-if="currentUser.confirmed === true"
          class="ma-2"
          color="success"
          variant="text"
          prepend-icon="mdi-checkbox-marked-circle"
        >
          Confirmed
        </v-chip>
        <v-chip
          v-else
          class="ma-2"
          color="warning"
          variant="text"
          prepend-icon="mdi-alert-circle"
        >
          Not confirmed
        </v-chip>
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
import useSnackbarStore from "@admin/store/modules/snackbar";
import useUsersStore from "@admin/store/modules/users";
import { IUser } from "@admin/interfaces/IUser";
import useAuthStore from "@admin/store/modules/auth";
import UserDelete from "../components/User/UserDelete.vue";
import { INotificationsError } from "../interfaces/INotifications";

const route = useRoute();
const userStore = useUsersStore();
const snackbarStore = useSnackbarStore();
const authStore = useAuthStore();

const userId = computed(() => route.params.id as string);

onBeforeMount(async () => {
  try {
    await userStore.get(userId.value);
  } catch {
    snackbarStore.showSnackbarErrorAction(INotificationsError.userDetails);
  }
});

const currentUser = computed(() => userStore.getUser as IUser);

const loginToken = async () => {
  try {
    const token = await authStore.loginToken(currentUser.value);

    const url = `/login?token=${token}`;
    window.open(url, "_target");
  } catch {
    snackbarStore.showSnackbarErrorAction(INotificationsError.errorLoginToken);
  }
};
</script>
