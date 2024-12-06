<template>
  <v-table class="bg-background border rounded" data-test="device-tunnels-table">
    <thead class="bg-v-theme-background">
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          class="text-center"
          :data-test="`device-tunnels-header-${i}`"
        >
          <span>{{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="tunnelList.length > 0" data-test="device-tunnels-rows">
      <slot name="rows">
        <tr
          v-for="(tunnel, i) in tunnelList"
          :key="i"
          :data-test="`device-tunnel-row-${i}`"
        >
          <td class="text-center" data-test="device-tunnel-url">
            <a
              :href="`${urlProtocol}//${tunnel.address}.${url}`"
              target="_blank"
              rel="noopener noreferrer"
              :data-test="`device-tunnel-link-${i}`"
            >
              {{ `${urlProtocol}//${tunnel.address}.${url}` }}
            </a>
          </td>

          <td class="text-center" data-test="device-tunnel-host">
            {{ tunnel.host }}
          </td>

          <td class="text-center" data-test="device-tunnel-port">
            {{ tunnel.port }}
          </td>

          <td class="text-center" data-test="device-tunnel-actions">
            <TunnelDelete
              :uid="tunnel.device"
              :address="tunnel.address"
              @update="getTunnels()"
              :data-test="`device-tunnel-delete-${i}`"
            />
          </td>
        </tr>
      </slot>
    </tbody>
    <div v-else class="pa-4 text-subtitle-2" data-test="device-tunnels-empty">
      <p>No data available</p>
    </div>
  </v-table>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { useStore } from "@/store";
import TunnelDelete from "./TunnelDelete.vue";
import { envVariables } from "@/envVariables";

const store = useStore();
const route = useRoute();
const tunnelList = computed(() => store.getters["tunnels/listTunnels"]);
const deviceId = computed(() => route.params.id);
const url = ref(window.location.host);
const urlProtocol = ref(window.location.protocol);

const getTunnels = async () => {
  await store.dispatch("tunnels/get", deviceId.value);
};

onMounted(() => {
  if (envVariables.isEnterprise && envVariables.hasTunnels) {
    getTunnels();
  }
});

const headers = [
  {
    text: "Address",
    value: "address",
  },
  {
    text: "Host",
    value: "host",
  },
  {
    text: "Port",
    value: "port",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

defineExpose({ getTunnels });
</script>
