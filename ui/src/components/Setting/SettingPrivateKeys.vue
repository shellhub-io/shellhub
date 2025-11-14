<template>
  <PrivateKeyAdd
    v-model="privateKeyAdd"
    @update="getPrivateKeys"
  />
  <v-container fluid>
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="card"
    >
      <v-row cols="12">
        <v-col cols="6">
          <v-card-item class="pa-0 ma-0 mb-2">
            <v-list-item data-test="card-header">
              <template #title>
                <h1 data-test="card-title">Private Keys</h1>
              </template>
              <template #subtitle>
                <span data-test="card-subtitle">
                  Manage your private keys securely with ShellHub
                </span>
              </template>
            </v-list-item>
          </v-card-item>
        </v-col>
        <v-col
          cols="6"
          class="d-flex justify-end"
        >
          <v-btn
            color="primary"
            variant="elevated"
            data-test="card-button"
            @click="privateKeyAdd = true"
          >
            Add Private Key
          </v-btn>
        </v-col>
      </v-row>
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
