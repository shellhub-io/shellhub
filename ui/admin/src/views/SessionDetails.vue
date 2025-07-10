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
            @click="session.device?.uid && goToDevice(session.device.uid)"
            @keyup="session.device?.uid && goToDevice(session.device.uid)"
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

<script setup lang="ts">
import { computed, ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { IAdminSession } from "@admin/interfaces/ISession";
import useSessionsStore from "@admin/store/modules/sessions";
import useSnackbar from "@/helpers/snackbar";

const route = useRoute();
const router = useRouter();
const snackbar = useSnackbar();
const sessionStore = useSessionsStore();

const session = ref({} as IAdminSession);
const sessionId = computed(() => route.params.id);

onMounted(async () => {
  try {
    await sessionStore.get(sessionId.value as string);
    session.value = sessionStore.getSession;
  } catch {
    snackbar.showError("Failed to get session details.");
  }
});

const goToDevice = (deviceId: string) => {
  router.push({ name: "deviceDetails", params: { id: deviceId } });
};

const sessionIsEmpty = computed(() => sessionStore.getSession && sessionStore.getSession.device_uid?.length === 0);

defineExpose({ session });
</script>

<style scoped>
.link {
  text-decoration: underline;
  cursor: pointer;
}
</style>
