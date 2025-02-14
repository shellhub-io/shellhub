<template>
  <div class="d-flex pa-0 align-center">
    <h1>Session Details</h1>
  </div>

  <v-card class="mt-2 pa-4" v-if="!sessionIsEmpty">
    <v-card-text>
      <div>
        <div class="text-overline mt-3">
          <h3>uid:</h3>
        </div>
        <div :data-test="session.uid">
          <p>{{ session.uid }}</p>
        </div>
      </div>

      <div v-if="session.device">
        <div class="text-overline mt-3">
          <h3>Device uid:</h3>
        </div>
        <div :data-test="session.device.uid">
          <p
            @click="goToDevice(session.device.uid)"
            @keyup="goToDevice(session.device.uid)"
            tabindex="0"
            class="link"
          >
            {{ session.device.uid }}
          </p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Tenant uid:</h3>
        </div>
        <div :data-test="session.tenant_id">
          <p>{{ session.tenant_id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Username</h3>
        </div>
        <div :data-test="session.username">
          <p>{{ session.username }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Ip Adress:</h3>
        </div>
        <div :data-test="session.ip_address">
          <p>{{ session.ip_address }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Last Seen:</h3>
        </div>
        <div :data-test="session.last_seen">
          <p>{{ session.last_seen }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Active:</h3>
        </div>
        <div :data-test="session.active">
          <p>{{ session.active }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Terminal:</h3>
        </div>
        <div :data-test="session.term">
          <p>{{ session.term }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Type:</h3>
        </div>
        <div :data-test="session.type">
          <p>{{ session.type }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
  <v-card class="mt-2 pa-4" v-else>
    <p class="text-center">Something is wrong, try again !</p>
  </v-card>
</template>

<script lang="ts">
import { computed, ref, defineComponent, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { INotificationsError } from "../interfaces/INotifications";
import { ISessions } from "../interfaces/ISession";
import { useStore } from "../store";

export default defineComponent({
  setup() {
    const store = useStore();
    const route = useRoute();
    const router = useRouter();

    const session = ref({} as ISessions);
    const sessionId = computed(() => route.params.id);

    onMounted(async () => {
      try {
        await store.dispatch("sessions/get", sessionId.value);
        session.value = store.getters["sessions/session"];
      } catch {
        store.dispatch("snackbar/showSnackbarErrorAction", INotificationsError.sessionDetails);
      }
    });

    const goToDevice = (deviceId: string) => {
      router.push({ name: "deviceDetails", params: { id: deviceId } });
    };

    const sessionIsEmpty = computed(() => store.getters["sessions/get"] && store.getters["sessions/get"].lenght === 0);

    return {
      session,
      goToDevice,
      sessionIsEmpty,
    };
  },
});
</script>

<style scoped>
.link {
  text-decoration: underline;
  cursor: pointer;
}
</style>
