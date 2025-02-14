<template>
  <div class="d-flex pa-0 align-center">
    <h1>Namespace Details</h1>
  </div>
  <v-card class="mt-2 pa-4">
    <v-card-text>
      <div>
        <div class="text-overline mt-3">
          <h3>name:</h3>
        </div>
        <div :data-test="namespace.name">
          <p>{{ namespace.name }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Devices:</h3>
        </div>
        <div :data-test="namespace.devices_count">
          <p>{{ namespace.devices_count || 0 }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Owner:</h3>
        </div>
        <div :data-test="namespace.owner">
          <p
            @click="goToUser(namespace.owner)"
            @keyup="goToUser(namespace.owner)"
            tabindex="0"
            class="link"
          >
            {{ namespace.owner }}
          </p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Tenant Id:</h3>
        </div>
        <div :data-test="namespace.tenant_id">
          <p>{{ namespace.tenant_id }}</p>
        </div>
      </div>

      <div>
        <div class="text-overline mt-3">
          <h3>Members:</h3>
        </div>
        <ul v-for="(member, index) in namespace.members" :key="index">
          <li
            class="ml-8"
            v-for="(value, name, index) in member"
            :key="index"
          >
            <div v-if="name === 'id'">
              <span class="font-weight-bold mr-1" :data-test="name">{{ name }}:</span>
              <span
                @click="goToUser(namespace.owner)"
                @keyup="goToUser(namespace.owner)"
                tabindex="0"
                class="link field-value"
                :data-test="value"
              >{{ value }}</span
              >
            </div>
            <div v-else>
              <span class="font-weight-bold mr-1" :data-test="name">{{ name }}:</span>
              <span :data-test="value" class="field-value">{{ value }}</span>
            </div>
          </li>
        </ul>
      </div>

      <div v-if="namespace.settings">
        <div class="text-overline mt-3">
          <h3>Session Record:</h3>
        </div>
        <div :data-test="namespace.settings.session_record">
          <p>{{ namespace.settings.session_record }}</p>
        </div>
      </div>
    </v-card-text>
  </v-card>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { INamespace } from "../interfaces/INamespace";
import { useStore } from "../store";

export default defineComponent({
  setup() {
    const store = useStore();
    const route = useRoute();
    const router = useRouter();
    const loading = ref(false);
    const namespace = ref({} as INamespace);

    const namespaceId = computed(() => route.params.id);

    onMounted(async () => {
      loading.value = true;
      await store.dispatch("namespaces/get", namespaceId.value);
      namespace.value = store.getters["namespaces/get"];
      loading.value = false;
    });

    const goToUser = (userId: string) => {
      router.push({ name: "userDetails", params: { id: userId } });
    };

    return {
      namespace,
      goToUser,
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
