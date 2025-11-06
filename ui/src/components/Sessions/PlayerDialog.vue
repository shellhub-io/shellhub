<template>
  <BaseDialog
    v-model="showDialog"
    :transition="false"
    :force-fullscreen="true"
    :scrollable="false"
  >
    <div class="bg-v-theme-surface position-relative w-100 fill-height overflow-hidden">
      <v-btn
        class="position-absolute top-0 right-0 ma-2 close-btn"
        variant="text"
        data-test="close-btn"
        icon="mdi-close"
        @click="showDialog = false"
      />

      <Player
        :logs
        @close="showDialog = false"
      />
    </div>
  </BaseDialog>
</template>

<script setup lang="ts">
import BaseDialog from "@/components/Dialogs/BaseDialog.vue";
import Player from "./Player.vue";

const { logs } = defineProps<{ logs: string | null }>();

const showDialog = defineModel({ required: true, type: Boolean });
</script>

<style lang="scss" scoped>
.close-btn { z-index: 999; }
</style>
