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
      <CopyWarning ref="copyWarningRef" :copied-item="'API Key'">
        <template #default="{ copyText }">
          <v-text-field
            :value="apiKey"
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
                @click="copyText(apiKey)"
                data-test="copy-key-icon-btn"
              />
            </template>
          </v-text-field>
        </template>
      </CopyWarning>
    </div>
  </MessageDialog>
</template>

<script setup lang="ts">
import { ref } from "vue";
import MessageDialog from "@/components/MessageDialog.vue";
import CopyWarning from "@/components/User/CopyWarning.vue";

const props = defineProps<{ apiKey: string }>();
const showDialog = defineModel<boolean>({ required: true });
const copyWarningRef = ref<InstanceType<typeof CopyWarning>>();

const copyKey = async () => {
  if (copyWarningRef.value?.copyFn) {
    await copyWarningRef.value.copyFn(props.apiKey);
  }
};

const close = () => {
  showDialog.value = false;
};

defineExpose({ showDialog, close });
</script>

<style scoped>
.monospace-field :deep(.v-field__input) {
  font-family: 'Roboto Mono', monospace;
  font-size: 14px;
}
</style>
