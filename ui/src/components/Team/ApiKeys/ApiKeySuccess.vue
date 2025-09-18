<template>
  <MessageDialog
    v-model="showDialog"
    @close="close"
    title="API Key created successfully!"
    description="Make sure to copy your key now as you will not be able to see it again."
    icon="mdi-check-circle"
    icon-color="success"
    confirm-text="Copy to clipboard"
    cancel-text="Close"
    cancel-data-test="close-btn"
    confirm-data-test="copy-btn"
    @confirm="copyKey"
    @cancel="close"
    data-test="api-key-success-dialog"
  >
    <div class="px-4">
      <v-text-field
        :model-value="apiKey || ''"
        variant="outlined"
        readonly
        density="compact"
        class="monospace-field"
        data-test="generated-key-field"
      >
        <template #append-inner>
          <v-btn
            icon="mdi-content-copy"
            color="primary"
            variant="text"
            size="small"
            @click="copyKey"
            data-test="copy-key-icon-btn"
          />
        </template>
      </v-text-field>
    </div>
  </MessageDialog>
</template>

<script setup lang="ts">
import { } from "vue";
import MessageDialog from "@/components/MessageDialog.vue";
import useSnackbar from "@/helpers/snackbar";

interface Props {
  apiKey?: string;
  keyName?: string;
}

const props = withDefaults(defineProps<Props>(), {
  apiKey: "",
  keyName: "",
});
const snackbar = useSnackbar();
const showDialog = defineModel<boolean>({ required: true });

const copyKey = async () => {
  try {
    await navigator.clipboard.writeText(props.apiKey || "");
    snackbar.showSuccess("API Key copied to clipboard!");
  } catch (err) {
    console.error("Failed to copy: ", err);
    snackbar.showError("Failed to copy API key to clipboard.");
  }
};

const close = () => {
  showDialog.value = false;
};
</script>

<style scoped>
.monospace-field :deep(.v-field__input) {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
}
</style>
