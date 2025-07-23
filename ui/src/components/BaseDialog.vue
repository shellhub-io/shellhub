<template>
  <v-dialog
    v-model="showDialog"
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
  /* Breakpoints
  * - sm: 600px (default)
  * - md: 960px
  * - lg: 1280px
  * - xl: 1920px
  * - xxl: 2560px
  */
  breakpoint?: "sm" | "md" | "lg" | "xl" | "xxl"
  forceFullscreen?: boolean
}>();

const showDialog = defineModel<boolean>({ required: true });
const { smAndDown, thresholds } = useDisplay();
const fullscreen = computed(() => props.forceFullscreen || smAndDown.value);
const maxWidth = computed(() => fullscreen.value ? undefined : thresholds.value[props.breakpoint || "sm"]);

defineExpose({ fullscreen, maxWidth });
</script>
