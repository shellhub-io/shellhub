<template>
  <div class="d-flex pa-0 align-center" data-test="title">
    <h1>API Keys</h1>

    <v-spacer />

    <div class="d-flex" data-test="api-key-generate">
      <ApiKeyGenerate @update="refreshApiKeys" />
    </div>
  </div>

  <div class="mt-2" data-test="api-key-list">
    <ApiKeyList ref="apiKeyList" />
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
import ApiKeyList from "@/components/Team/ApiKeys/ApiKeyList.vue";

const apiKeyList = ref();

const refreshApiKeys = () => {
  apiKeyList.value.refresh();
};
</script>
