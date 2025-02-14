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

      <UserDelete :id="currentUser.id" redirect />
    </v-col>
  </div>

  <v-card v-if="!currentUserIsEmpty" class="mt-2 pa-4">
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

<script lang="ts">
import { computed, defineComponent, onMounted } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "../store";
import UserDelete from "../components/User/UserDelete.vue";
import { INotificationsError } from "../interfaces/INotifications";

export default defineComponent({
  setup() {
    const store = useStore();
    const route = useRoute();

    const userId = computed(() => route.params.id as string);

    onMounted(async () => {
      try {
        await store.dispatch("users/get", userId.value);
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.userDetails);
      }
    });

    const currentUser = computed(() => store.getters["users/user"]);
    const currentUserIsEmpty = computed(() => store.getters["users/user"] && store.getters["users/user"].lenght === 0);

    const loginToken = async () => {
      try {
        const token = await store.dispatch("auth/loginToken", currentUser.value);

        const url = `/login?token=${token}`;
        window.open(url, "_target");
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.errorLoginToken);
      }
    };

    return {
      userId,
      currentUser,
      headers: [
        {
          text: "Id",
          value: "id",
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
          value: "ownedNs",
          align: "center",
        },
      ],
      loginToken,
      currentUserIsEmpty,
    };
  },
  components: { UserDelete },
});
</script>
