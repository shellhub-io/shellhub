<template>
  <v-container fluid>
    <PrivateKeyAdd v-model="privateKeyAdd" @update="getPrivateKeys" />
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0 ma-0 mb-2"
          data-test="card-header"
        >
          <template v-slot:title>
            <h1 data-test="card-title">Private Keys</h1>
          </template>
          <template v-slot:subtitle>
            <span data-test="card-subtitle">
              Manage your private keys securely with ShellHub
            </span>
          </template>
          <template v-slot:append>
            <v-btn
              @click="privateKeyAdd = true"
              color="primary"
              variant="text"
              class="bg-secondary border"
              data-test="card-button"
            >Add Private Key</v-btn>
          </template>
        </v-list-item>

      </v-card-item>

      <PrivateKeyList data-test="private-key-list" />
    </v-card>

  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "../PrivateKeys/PrivateKeyList.vue";
import handleError from "@/utils/handleError";
import usePrivateKeysStore from "@/store/modules/private_keys";

const privateKeysStore = usePrivateKeysStore();
const privateKeyAdd = ref(false);

const getPrivateKeys = () => {
  try {
    privateKeysStore.getPrivateKeyList();
  } catch (error: unknown) {
    handleError(error);
  }
};
</script>
