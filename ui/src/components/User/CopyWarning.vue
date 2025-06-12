<template>
  <div>
    <slot :copyText="handleCopy" />

    <v-dialog v-model="showDialog" width="500">
      <v-card class="bg-grey-darken-4 bg-v-theme-surface">
        <v-card-title class="text-h5 pa-5 bg-primary">Copying is not allowed</v-card-title>
        <v-card-text>
          Clipboard access is only permitted on secure (HTTPS) or localhost origins.
          Please ensure your instance is secure to enable clipboard features.
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="primary" @click="showDialog = false">OK</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useClipboard, useMagicKeys } from "@vueuse/core";

const props = defineProps<{
  onSuccess:() => void;
  macro?: string;
}>();

const showDialog = ref(false);
const { copy } = useClipboard();

const handleCopy = async (text: string) => {
  const isSecure = globalThis?.location?.protocol === "https:" || globalThis?.location?.hostname === "localhost";

  if (!isSecure) {
    showDialog.value = true;
    return;
  }

  try {
    await copy(text);
    props.onSuccess();
  } catch (error) {
    showDialog.value = true;
  }
};

onMounted(() => {
  if (!props.macro) return;

  let executed = false;

  useMagicKeys({
    passive: false,
    onEventFired(e) {
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

defineExpose({
  copyFn: handleCopy,
});
</script>
