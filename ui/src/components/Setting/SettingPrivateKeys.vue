<template>
  <v-container fluid>
    <PrivateKeyAdd
      v-model="privateKeyAdd"
      @update="getPrivateKeys"
    />
    <PageHeader
      icon="mdi-shield-key"
      title="Private Keys"
      overline="Settings"
      description="Securely store and manage your SSH private keys for automatic device authentication."
      icon-color="primary"
      data-test="card"
    >
      <template
        v-if="hasPrivateKeys"
        #actions
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
    </PageHeader>

    <v-card
      v-if="hasPrivateKeys"
      variant="flat"
      class="bg-transparent"
    >
      <PrivateKeyList
        v-if="hasPrivateKeys"
        data-test="private-key-list"
      />
    </v-card>

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
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "../PrivateKeys/PrivateKeyList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import PageHeader from "../PageHeader.vue";
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
