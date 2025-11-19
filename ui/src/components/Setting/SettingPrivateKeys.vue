<template>
  <v-container fluid>
    <PrivateKeyAdd
      v-model="privateKeyAdd"
      @update="getPrivateKeys"
    />
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
          <template #title>
            <h1 data-test="card-title">
              Private Keys
            </h1>
          </template>
          <template #subtitle>
            <span data-test="card-subtitle">
              Manage your private keys securely with ShellHub
            </span>
          </template>
          <template
            v-if="hasPrivateKeys"
            #append
          >
            <v-btn
              color="primary"
              variant="elevated"
              data-test="card-button"
              @click="privateKeyAdd = true"
            >
              Add Private Key
            </v-btn>
          </template>
        </v-list-item>
      </v-card-item>

      <PrivateKeyList
        v-if="hasPrivateKeys"
        data-test="private-key-list"
      />

      <NoItemsMessage
        v-else
        item="Private Keys"
        icon="mdi-shield-key"
        data-test="no-items-message-component"
      >
        <template #content>
          <p>
            ShellHub provides secure storage for your SSH private keys.
            This allows you to authenticate with your devices securely and automatically.
          </p>
          <p>
            By adding your Private Keys here, you can streamline access to your
            devices without managing passwords manually for every connection.
          </p>
        </template>
        <template #action>
          <v-btn
            color="primary"
            variant="elevated"
            @click="privateKeyAdd = true"
          >
            Add Private Key
          </v-btn>
        </template>
      </NoItemsMessage>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "../PrivateKeys/PrivateKeyList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import handleError from "@/utils/handleError";
import usePrivateKeysStore from "@/store/modules/private_keys";

const privateKeysStore = usePrivateKeysStore();
const privateKeyAdd = ref(false);

const hasPrivateKeys = computed(() => privateKeysStore.privateKeys.length > 0);

const getPrivateKeys = () => {
  try {
    privateKeysStore.getPrivateKeyList();
  } catch (error: unknown) {
    handleError(error);
  }
};
</script>
