<template>
  <v-container fluid class="ml-0 pa-0" max-width="960">
    <PrivateKeyAdd v-model="privateKeyAdd" @update="getPrivateKeys" />
    <v-card
      variant="flat"
      class="bg-transparent"
      data-test="account-profile-card"
    >
      <v-card-item>
        <v-list-item
          class="pa-0 ma-0 mb-2"
          data-test="profile-header"
        >
          <template v-slot:title>
            <h1>Tags</h1>
          </template>
          <template v-slot:subtitle>
            <span data-test="profile-subtitle">Manage your device and connector tags</span>
          </template>
        </v-list-item>
      </v-card-item>
      <TagList />
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from "vue";
import PrivateKeyAdd from "../PrivateKeys/PrivateKeyAdd.vue";
import TagList from "../Tags/TagList.vue";
import { useStore } from "@/store";
import handleError from "@/utils/handleError";

const store = useStore();
const privateKeyAdd = ref(false);

const getPrivateKeys = async () => {
  try {
    await store.dispatch("privateKey/fetch");
  } catch (error: unknown) {
    handleError(error);
  }
};
</script>
