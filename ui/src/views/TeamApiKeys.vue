<template>
  <PageHeader
    icon="mdi-key"
    title="API Keys"
    overline="Team Management"
    description="API Keys allow you to authenticate and integrate external applications or scripts with ShellHub securely."
    icon-color="primary"
    data-test="api-key-title"
  >
    <template #actions>
      <ApiKeyGenerate
        v-if="hasApiKeys"
        @update="refreshApiKeys"
      />
    </template>
  </PageHeader>

  <div
    v-if="loading"
    class="d-flex justify-center mt-4"
  >
    <v-progress-circular
      indeterminate
      color="primary"
    />
  </div>

  <div
    v-else-if="hasApiKeys"
    class="mt-2"
  >
    <ApiKeyList
      ref="apiKeyList"
      data-test="api-key-list"
    />
  </div>

  <NoItemsMessage
    v-else
    class="mt-2"
    item="API Keys"
    icon="mdi-cloud-key"
    data-test="no-items-message-component"
  >
    <template #content>
      <p>
        API Keys allow you to authenticate and integrate external applications
        or scripts with ShellHub securely.
      </p>
      <p>
        They are essential for automating tasks without manual user intervention,
        enabling third-party apps to interact with your resources based on specific permissions.
      </p>
    </template>
    <template #action>
      <ApiKeyGenerate @update="refreshApiKeys" />
    </template>
  </NoItemsMessage>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import ApiKeyGenerate from "@/components/Team/ApiKeys/ApiKeyGenerate.vue";
import ApiKeyList from "@/components/Team/ApiKeys/ApiKeyList.vue";
import NoItemsMessage from "@/components/NoItemsMessage.vue";
import PageHeader from "@/components/PageHeader.vue";
import useApiKeysStore from "@/store/modules/api_keys";
import handleError from "@/utils/handleError";
import useSnackbar from "@/helpers/snackbar";

const apiKeyList = ref<InstanceType<typeof ApiKeyList>>();
const apiKeysStore = useApiKeysStore();
const snackbar = useSnackbar();
const loading = ref(true);

const hasApiKeys = computed(() => apiKeysStore.apiKeysCount > 0);

const fetchInitialKeys = async () => {
  try {
    loading.value = true;
    await apiKeysStore.fetchApiKeys({
      page: 1,
      perPage: 10,
    });
  } catch (error) {
    snackbar.showError("Failed to load API keys.");
    handleError(error);
  } finally {
    loading.value = false;
  }
};

const refreshApiKeys = async () => {
  if (apiKeyList.value) {
    await apiKeyList.value.refresh();
  } else {
    await fetchInitialKeys();
  }
};

onMounted(async () => {
  await fetchInitialKeys();
});
</script>
