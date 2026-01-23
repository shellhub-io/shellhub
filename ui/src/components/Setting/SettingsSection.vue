<template>
  <v-card
    variant="flat"
    class="bg-transparent"
  >
    <v-card-text class="pt-0">
      <v-list
        v-bind="attrs"
        border
        rounded
        class="bg-background pa-0"
      >
        <slot />
      </v-list>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { useAttrs } from "vue";

defineOptions({ inheritAttrs: false });

const attrs = useAttrs();
</script>
