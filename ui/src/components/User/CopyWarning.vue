<template>
  <div>
    <slot :copyText="handleCopy" />

    <MessageDialog
      v-model="showDialog"
      @close="showDialog = false"
      @cancel="showDialog = false"
      title="Copying is not allowed"
      description="Clipboard access is only permitted on secure (HTTPS) or localhost origins.
       Please ensure your instance is secure to enable clipboard features."
      icon="mdi-alert"
      icon-color="warning"
      confirm-text="OK"
      confirm-color="primary"
      confirm-data-test="copy-warning-ok-btn"
      data-test="copy-warning-dialog"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useClipboard, useMagicKeys } from "@vueuse/core";
import useSnackbar from "@/helpers/snackbar";
import MessageDialog from "../MessageDialog.vue";

const props = defineProps<{
  macro?: string;
  copiedItem?: string;
  bypass?: boolean;
}>();

const snackbar = useSnackbar();
const showDialog = ref(false);
const { copy } = useClipboard();

const handleCopy = async (text: string) => {
  if (props.bypass) return;

  const isSecure = globalThis?.isSecureContext;
  if (!isSecure) {
    showDialog.value = true;
    return;
  }

  try {
    await copy(text);
    if (props.copiedItem) {
      snackbar.showInfo(`${props.copiedItem} copied to clipboard!`);
    } else {
      snackbar.showInfo("Successfully copied to clipboard!");
    }
  } catch {
    showDialog.value = true;
  }
};

onMounted(() => {
  if (!props.macro) return;
  let executed = false;

  useMagicKeys({
    passive: false,
    onEventFired(e) {
      if (props.bypass) return;

      if (!executed && e.ctrlKey && e.key === "c" && e.type === "keydown") {
        executed = true;
        handleCopy(props.macro as string);
        e.preventDefault();
        setTimeout(() => {
          executed = false;
        }, 500);
      }
    },
  });
});

defineExpose({ copyFn: handleCopy });
</script>
