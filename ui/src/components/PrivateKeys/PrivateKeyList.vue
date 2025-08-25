<template>
  <v-table class="bg-background border rounded mx-4">
    <thead class="bg-v-theme-background" data-test="privateKey-thead">
      <tr>
        <th
          v-for="(head, i) in headers"
          :key="i"
          :class="head.align ? `text-${head.align}` : 'text-center'"
        >
          <span> {{ head.text }}</span>
        </th>
      </tr>
    </thead>
    <tbody v-if="privateKeys.length">
      <tr v-for="(privateKey, i) in privateKeys" :key="i">
        <td class="text-center" data-test="privateKey-name">
          {{ privateKey.name }}
        </td>
        <td class="text-center" data-test="privateKey-fingerprint">
          {{ getKeyFingerprint(privateKey) }}
        </td>
        <td class="text-center">
          <v-menu
            location="bottom"
            scrim
            eager
          >
            <template v-slot:activator="{ props }">
              <v-btn
                v-bind="props"
                variant="plain"
                class="border rounded bg-v-theme-background"
                density="comfortable"
                size="default"
                icon="mdi-format-list-bulleted"
                data-test="privateKey-actions"
              />
            </template>
            <v-list class="bg-v-theme-surface" lines="two" density="compact">
              <PrivateKeyEdit
                :private-key="privateKey"
                @update="getPrivateKeys"
              />

              <PrivateKeyDelete
                :id="privateKey.id"
                @update="getPrivateKeys"
              />
            </v-list>
          </v-menu>
        </td>
      </tr>
    </tbody>
    <div v-else sm="12" class="text-start mt-2 mb-3" data-test="no-private-key-warning">
      <span class="ml-4">No data available</span>
    </div>
  </v-table>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import PrivateKeyDelete from "./PrivateKeyDelete.vue";
import PrivateKeyEdit from "./PrivateKeyEdit.vue";
import handleError from "@/utils/handleError";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { convertToFingerprint } from "@/utils/validate";
import usePrivateKeysStore from "@/store/modules/private_keys";

const privateKeysStore = usePrivateKeysStore();

const headers = [
  {
    text: "Name",
    value: "name",
    align: "center",
    sortable: true,
  },
  {
    text: "Fingerprint",
    value: "data",
    align: "center",
    sortable: true,
  },
  {
    text: "Actions",
    value: "actions",
    align: "center",
    sortable: false,
  },
];
const privateKeys = computed(() => privateKeysStore.privateKeys);

const getPrivateKeys = () => {
  try {
    privateKeysStore.getPrivateKeyList();
  } catch (error: unknown) {
    handleError(error);
  }
};

const getKeyFingerprint = (privateKey: IPrivateKey) => {
  if (privateKey.fingerprint) {
    return privateKey.fingerprint;
  }

  const fingerprint = convertToFingerprint(privateKey.data);
  return fingerprint || "Fingerprint not available";
};

onMounted(() => {
  getPrivateKeys();
});
</script>
