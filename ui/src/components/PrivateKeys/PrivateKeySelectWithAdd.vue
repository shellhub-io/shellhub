<template>
  <div>
    <v-select
      v-model="selectedPrivateKeyName"
      :items="privateKeysNames"
      :list-props="{ class: 'py-0' }"
      label="Private Key"
      hint="Select a private key file for authentication"
      persistent-hint
      data-test="private-keys-select"
    >
      <template #append-item>
        <v-divider />
        <v-list-item
          data-test="add-private-key-btn"
          @click="showPrivateKeyAdd = true"
        >
          <template #prepend>
            <v-icon
              color="primary"
              icon="mdi-plus"
            />
          </template>
          <v-list-item-title class="text-primary">
            Add New Private Key
          </v-list-item-title>
        </v-list-item>
      </template>
    </v-select>

    <PrivateKeyAdd
      v-model="showPrivateKeyAdd"
      @update="handlePrivateKeyAdded"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import PrivateKeyAdd from "@/components/PrivateKeys/PrivateKeyAdd.vue";
import usePrivateKeysStore from "@/store/modules/private_keys";
import { IPrivateKey } from "@/interfaces/IPrivateKey";

const emit = defineEmits<{ "key-added": [] }>();

const selectedPrivateKeyName = defineModel<string>({ required: true });
const privateKeysStore = usePrivateKeysStore();
const showPrivateKeyAdd = ref(false);

const privateKeysNames = computed(() => privateKeysStore.privateKeys.map((item: IPrivateKey) => item.name));

const handlePrivateKeyAdded = () => {
  privateKeysStore.getPrivateKeyList();
  const newestKey = privateKeysStore.privateKeys[privateKeysStore.privateKeys.length - 1];
  if (newestKey) {
    selectedPrivateKeyName.value = newestKey.name;
    emit("key-added");
  }
};

defineExpose({ selectedPrivateKeyName });
</script>
