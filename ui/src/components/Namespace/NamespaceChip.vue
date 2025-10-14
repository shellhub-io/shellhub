<template>
  <v-chip label color="primary" class="text-uppercase">
    {{ firstNamespaceLetter }}
  </v-chip>
</template>

<script setup lang="ts">
import { computed } from "vue";

interface Props {
  name?: string;
}

const props = withDefaults(defineProps<Props>(), {
  name: "",
});

const firstNamespaceLetter = computed(() => props.name.charAt(0) || "");
</script>
