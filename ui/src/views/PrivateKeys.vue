<template>
  <PrivateKeyAdd
    v-model="showPrivateKeyAdd"
    @update="getPrivateKeys"
  />
  <PageHeader
    icon="mdi-shield-key"
    title="Private Keys"
    overline="Settings"
    description="Securely store and manage your SSH private keys for automatic device authentication."
    icon-color="primary"
    data-test="private-keys-page-header"
  >
    <template
      v-if="hasPrivateKeys"
      #actions
    >
      <v-btn
        color="primary"
        variant="elevated"
        data-test="add-private-key-btn"
        text="Add Private Key"
        @click="showPrivateKeyAdd = true"
      />
    </template>
  </PageHeader>

  <PrivateKeyList
    v-if="hasPrivateKeys"
    data-test="private-keys-list"
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
        data-test="no-items-add-private-key-btn"
        text="Add Private Key"
        @click="showPrivateKeyAdd = true"
      />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import PrivateKeyList from "@/components/PrivateKeys/PrivateKeyList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";
import PageHeader from "@/components/PageHeader.vue";

const privateKeysStore = usePrivateKeysStore();
const showPrivateKeyAdd = ref(false);
const privateKeysList = ref<InstanceType<typeof PrivateKeyList> | null>(null);

const hasPrivateKeys = computed(() => privateKeysStore.privateKeys.length > 0);

const getPrivateKeys = () => { privateKeysList.value?.getPrivateKeysList(); };

defineExpose({ showPrivateKeyAdd });
</script>
