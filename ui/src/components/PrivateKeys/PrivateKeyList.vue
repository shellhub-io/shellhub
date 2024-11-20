<template>
  <v-table class="bg-v-theme-surface">
    <thead data-test="privateKey-thead">
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
    <tbody v-if="getListPrivateKeys.length">
      <tr v-for="(privateKey, i) in getListPrivateKeys" :key="i">
        <td class="text-center" data-test="privateKey-name">
          {{ privateKey.name }}
        </td>
        <td class="text-center" data-test="privateKey-fingerprint">
          {{ convertToFingerprint(privateKey.data) }}
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
                :keyObject="privateKey"
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
    <div v-else sm="12" class="text-start mt-2 text-medium-emphasis" data-test="no-private-key-warning">
      <span>No data avaliable</span>
    </div>
  </v-table>
</template>

<script setup lang="ts">
import { computed, onMounted } from "vue";
import { useStore } from "../../store";
import { convertToFingerprint } from "../../utils/validate";
import PrivateKeyDelete from "./PrivateKeyDelete.vue";
import PrivateKeyEdit from "./PrivateKeyEdit.vue";
import handleError from "@/utils/handleError";

const store = useStore();
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
const getListPrivateKeys = computed(() => store.getters["privateKey/list"]);

const getPrivateKeys = async () => {
  try {
    await store.dispatch("privateKey/fetch");
  } catch (error: unknown) {
    handleError(error);
  }
};

onMounted(() => {
  getPrivateKeys();
});
</script>
