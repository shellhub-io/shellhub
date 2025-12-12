<template>
  <div
    class="d-flex flex-column justify-space-between align-center flex-sm-row mb-2"
    data-test="private-keys-title"
  >
    <h1>Private Keys</h1>
    <v-btn
      color="primary"
      variant="elevated"
      data-test="private-key-add-btn"
      text="Add Private Key"
      @click="privateKeyAdd = true"
    />
  </div>

  <div data-test="private-keys-components">
    <PrivateKeyAdd
      v-model="privateKeyAdd"
      @update="getPrivateKeys"
    />

    <PrivateKeyList
      v-if="hasPrivateKeys"
      ref="privateKeysList"
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
          By adding your Private Keys here,
          you can streamline access to your devices without managing passwords manually for every connection.
        </p>
      </template>
      <template #action>
        <v-btn
          color="primary"
          variant="elevated"
          text="Add Private Key"
          @click="privateKeyAdd = true"
        />
      </template>
    </NoItemsMessage>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";

const privateKeysStore = usePrivateKeysStore();
const privateKeyAdd = ref(false);
const privateKeysList = ref<InstanceType<typeof PrivateKeyList> | null>(null);

const hasPrivateKeys = computed(() => privateKeysStore.privateKeys.length > 0);

const getPrivateKeys = () => { privateKeysList.value?.getPrivateKeysList(); };
</script>
