<template>
  <v-dialog
    v-model="showDialog"
    @update:model-value="handleModelValueChange"
    :fullscreen
    :max-width
  >
    <slot />
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useDisplay } from "vuetify";

const props = defineProps<{
  /* Thresholds
  * - sm: 600px (default)
  * - md: 960px
  * - lg: 1280px
  * - xl: 1920px
  * - xxl: 2560px
  */
  threshold?: "sm" | "md" | "lg" | "xl" | "xxl"
  forceFullscreen?: boolean
}>();

const emit = defineEmits(["close"]);
const showDialog = defineModel<boolean>({ required: true });
const { smAndDown, thresholds } = useDisplay();
const fullscreen = computed(() => props.forceFullscreen || smAndDown.value);
const maxWidth = computed(() => fullscreen.value ? undefined : thresholds.value[props.threshold || "sm"]);

const handleModelValueChange = (value: boolean) => { if (!value) emit("close"); };

defineExpose({ fullscreen, maxWidth });
</script>
