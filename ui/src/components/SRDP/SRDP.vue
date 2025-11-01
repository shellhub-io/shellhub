<template>
  <WindowDialog
    v-model="showDialog"
    :force-fullscreen="true"
    @close="close"
    title="Remote Desktop"
    :description="`Connected to ${deviceName || device}`"
    icon="mdi-monitor-screenshot"
    :show-close-button="true"
    :show-footer="false"
    :scrollable="false"
    class="bg-black h-100"
  >
    <div class="ma-0 pa-0 w-100 fill-height">
      <!-- Embed Ebitengine in iframe as recommended by Ebitengine docs -->
      <!-- This properly handles canvas, input events, and screen scaling -->
      <iframe 
        v-if="showDialog"
        :src="iframeUrl"
        class="SRDP-iframe"
        allow="autoplay"
        sandbox="allow-same-origin allow-scripts allow-pointer-lock"
        data-test="SRDP-iframe"
      />
    </div>
  </WindowDialog>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import WindowDialog from "@/components/Dialogs/WindowDialog.vue";

// Props - just need credentials to pass to iframe
const { device, username, password, display, deviceName } = defineProps<{
  device: string;
  username: string;
  password: string;
  display?: string;
  deviceName?: string;
}>();

const emit = defineEmits<{
  close: [];
}>();

const showDialog = defineModel<boolean>({ required: true });

// Construct iframe URL with credentials as query parameters
const iframeUrl = computed(() => {
  const params = new URLSearchParams({
    device,
    username,
    password,
  });
  if (display) {
    params.append('display', display);
  }
  return `/srdp.html?${params.toString()}`;
});

const close = () => {
  showDialog.value = false;
  emit("close");
};

defineExpose({ showDialog });
</script>

<style scoped lang="scss">
.SRDP-iframe {
  width: 100%;
  height: 100%;
  border: none;
  display: block;
  background-color: #000;
}
</style>
