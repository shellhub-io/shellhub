<template>
  <h1>Session Details</h1>
  <v-card class="mt-2 pa-4 bg-background border">
    <v-card-text v-if="!isSessionEmpty">
      <div>
        <h3 class="text-overline">uid:</h3>
        <p :data-test="session.uid">{{ session.uid }}</p>
      </div>

      <div v-if="session.device">
        <h3 class="text-overline mt-3">Device uid:</h3>
        <p
          :data-test="session.device.uid"
          @click="session.device?.uid && goToDevice(session.device.uid)"
          @keyup="session.device?.uid && goToDevice(session.device.uid)"
          tabindex="0"
          class="text-decoration-underline cursor-pointer"
        >
          {{ session.device.uid }}
        </p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Tenant uid:</h3>
        <p :data-test="session.tenant_id">{{ session.tenant_id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Username</h3>
        <p :data-test="session.username">{{ session.username }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Ip Address:</h3>
        <p :data-test="session.ip_address">{{ session.ip_address }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Last Seen:</h3>
        <p :data-test="session.last_seen">{{ session.last_seen }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Active:</h3>
        <p :data-test="session.active">{{ session.active }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Terminal:</h3>
        <p :data-test="session.term">{{ session.term }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Type:</h3>
        <p :data-test="session.type">{{ session.type }}</p>
      </div>
    </v-card-text>
    <p v-else class="text-center">Something is wrong, try again!</p>
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
const sessionId = computed(() => route.params.id);
const session = ref({} as IAdminSession);
const isSessionEmpty = computed(() => session.value && session.value.device_uid?.length === 0);

const goToDevice = (deviceId: string) => {
  router.push({ name: "deviceDetails", params: { id: deviceId } });
};

onMounted(async () => {
  try {
    session.value = await sessionStore.fetchSessionById(sessionId.value as string);
  } catch {
    snackbar.showError("Failed to get session details.");
  }
});

defineExpose({ session });
</script>
