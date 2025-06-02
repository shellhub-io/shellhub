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
          :class="formatKey(tunnel.expires_in) ? 'text-warning' : ''"
        >
          <td class="text-center" data-test="device-tunnel-url">
            <a
              :href="`${urlProtocol}//${tunnel.full_address}`"
              target="_blank"
              rel="noopener noreferrer"
              :data-test="`device-tunnel-link-${i}`"
              @click="handleTunnelLinkClick"
            >
              {{ `${urlProtocol}//${tunnel.full_address}` }}
            </a>
          </td>

          <td class="text-center" data-test="device-tunnel-host">
            {{ tunnel.host }}
          </td>

          <td class="text-center" data-test="device-tunnel-port">
            {{ tunnel.port }}
          </td>

          <td
            class="text-center"
            data-test="device-tunnel-expiration-date"
          >
            {{ formatDate(tunnel.expires_in) }}
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
import moment from "moment";
import { useStore } from "@/store";
import TunnelDelete from "./TunnelDelete.vue";
import { envVariables } from "@/envVariables";
import { ITunnel } from "@/interfaces/ITunnel";

const { deviceUid } = defineProps<{
  deviceUid: string;
}>();

const store = useStore();
const tunnelList = computed<Array<ITunnel>>(() => store.getters["tunnels/listTunnels"]);
const urlProtocol = ref(window.location.protocol);

const getTunnels = async () => {
  await store.dispatch("tunnels/get", deviceUid);
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
    text: "Expiration Date",
    value: "expires_in",
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const now = moment().utc();

const formatKey = (date: string) => {
  if (date === "0001-01-01T00:00:00Z") return false;

  const expiryDate = moment(date);

  return now.isAfter(expiryDate);
};

const formatDate = (expiresIn: string) => {
  if (expiresIn === "0001-01-01T00:00:00Z") return "Never Expires";

  const expirationDate = moment(expiresIn);
  const format = "MMM D YYYY, h:mm:ss";

  return now.isAfter(expirationDate)
    ? `Expired on ${expirationDate.format(format)}.`
    : `Expires on ${expirationDate.format(format)}.`;
};

const handleTunnelLinkClick = () => {
  setTimeout(() => {
    getTunnels();
  }, 30000);
};

defineExpose({ getTunnels });
</script>
