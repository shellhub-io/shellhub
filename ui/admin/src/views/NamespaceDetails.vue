<template>
  <div class="d-flex pa-0 align-center">
    <h1>Namespace Details</h1>
  </div>
  <v-card class="mt-2 pa-4">
    <v-card-text>
      <div>
        <h3 class="text-overline">Name:</h3>
        <p :data-test="namespace.name">{{ namespace.name }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Devices:</h3>
        <p data-test="namespace-devices-count">{{ sumDevicesCount(namespace) }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Owner:</h3>
        <p
          :data-test="namespace.owner"
          @click="goToUser(namespace.owner)"
          @keyup="goToUser(namespace.owner)"
          tabindex="0"
          class="text-decoration-underline cursor-pointer"
        >
          {{ namespace.owner }}
        </p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Tenant Id:</h3>
        <p :data-test="namespace.tenant_id">{{ namespace.tenant_id }}</p>
      </div>

      <div>
        <h3 class="text-overline mt-3">Members:</h3>
        <ul v-for="(member, index) in namespace.members" :key="index">
          <li
            class="ml-8"
            v-for="(value, name, index) in member"
            :key="index"
          >
            <div v-if="name === 'id'">
              <span class="font-weight-bold mr-1" :data-test="`${name}-item`">{{ name }}:</span>
              <span
                @click="goToUser(namespace.owner)"
                @keyup="goToUser(namespace.owner)"
                tabindex="0"
                class="text-decoration-underline cursor-pointer"
                :data-test="`${name}-value`"
              >{{ value }}</span
              >
            </div>
            <div v-else>
              <span class="font-weight-bold mr-1" :data-test="`${name}-item`">{{ name }}:</span>
              <span :data-test="`${name}-value`">{{ value }}</span>
            </div>
          </li>
        </ul>
      </div>

      <div v-if="namespace.settings">
        <h3 class="text-overline mt-3">Session Record:</h3>
        <p :data-test="namespace.settings.session_record">{{ namespace.settings.session_record }}</p>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import useNamespacesStore from "@admin/store/modules/namespaces";
import { IAdminNamespace } from "@admin/interfaces/INamespace";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

const namespacesStore = useNamespacesStore();
const snackbar = useSnackbar();
const route = useRoute();
const router = useRouter();
const loading = ref(false);
const namespace = ref({} as IAdminNamespace);

const namespaceId = computed(() => route.params.id);

onMounted(async () => {
  try {
    loading.value = true;
    namespace.value = await namespacesStore.fetchNamespaceById(namespaceId.value as string);
  } catch (error) {
    snackbar.showError("Failed to fetch namespace details.");
    handleError(error);
  }
  loading.value = false;
});

const goToUser = (userId: string) => {
  router.push({ name: "userDetails", params: { id: userId } });
};

const sumDevicesCount = (namespace: IAdminNamespace) => {
  const { devices_accepted_count: acceptedCount, devices_pending_count: pendingCount, devices_rejected_count: rejectedCount } = namespace;
  return (acceptedCount + pendingCount + rejectedCount) || 0;
};

defineExpose({ namespace });
</script>
