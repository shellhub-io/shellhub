<template>
  <DataTable
    v-model:page="page"
    v-model:items-per-page="itemsPerPage"
    :headers
    :items="privateKeys"
    :total-count="privateKeyCount"
    :loading
    :items-per-page-options="[10, 20, 50, 100]"
    data-test="private-keys-list"
  >
    <template #rows>
      <tr
        v-for="(privateKey, i) in privateKeys"
        :key="i"
        data-test="private-key-item"
      >
        <td
          class="text-center"
          data-test="private-key-name"
        >
          {{ privateKey.name }}
        </td>

        <td
          class="text-center"
          data-test="private-key-fingerprint"
        >
          {{ getKeyFingerprint(privateKey) }}
        </td>

        <td
          class="text-center"
          data-test="private-key-actions"
        >
          <v-menu
            location="bottom"
            scrim
            eager
          >
            <template #activator="{ props }">
              <v-btn
                v-bind="props"
                variant="plain"
                class="border rounded bg-v-theme-background"
                density="comfortable"
                size="default"
                icon="mdi-format-list-bulleted"
              />
            </template>
            <v-list
              class="bg-v-theme-surface"
              lines="two"
              density="compact"
            >
              <PrivateKeyEdit
                :private-key="privateKey"
                @update="getPrivateKeysList"
              />

              <PrivateKeyDelete
                :id="privateKey.id"
                @update="getPrivateKeysList"
              />
            </v-list>
          </v-menu>
        </td>
      </tr>
    </template>
  </DataTable>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted } from "vue";
import DataTable from "../Tables/DataTable.vue";
import PrivateKeyDelete from "./PrivateKeyDelete.vue";
import PrivateKeyEdit from "./PrivateKeyEdit.vue";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";
import { IPrivateKey } from "@/interfaces/IPrivateKey";
import { convertToFingerprint } from "@/utils/sshKeys";
import usePrivateKeysStore from "@/store/modules/private_keys";

const headers = [
  {
    text: "Name",
    value: "name",
    sortable: true,
  },
  {
    text: "Fingerprint",
    value: "data",
    sortable: true,
  },
  {
    text: "Actions",
    value: "actions",
  },
];

const privateKeysStore = usePrivateKeysStore();
const snackbar = useSnackbar();
const loading = ref(false);
const itemsPerPage = ref(10);
const page = ref(1);
const privateKeys = computed(() => privateKeysStore.privateKeys);
const privateKeyCount = computed(() => privateKeys.value.length);

const getPrivateKeysList = () => {
  try {
    loading.value = true;
    privateKeysStore.getPrivateKeyList();
  } catch (error: unknown) {
    snackbar.showError("Failed to load private keys.");
    handleError(error);
  }
  loading.value = false;
};

watch([page, itemsPerPage], () => { getPrivateKeysList(); });

const getKeyFingerprint = (privateKey: IPrivateKey) => {
  if (privateKey.fingerprint) return privateKey.fingerprint;

  const fingerprint = convertToFingerprint(privateKey.data);
  return fingerprint || "Fingerprint not available";
};

onMounted(() => { getPrivateKeysList(); });

defineExpose({ getPrivateKeysList });
</script>
