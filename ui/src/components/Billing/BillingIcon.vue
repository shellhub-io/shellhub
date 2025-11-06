<template>
  <v-icon
    :icon="iconString"
    size="1.5rem"
    :data-test="isDefaultIcon ? 'default-icon' : 'type-icon'"
  />
</template>

<script setup lang="ts">
import { computed } from "vue";

const { iconName } = defineProps<{ iconName: string }>();

const cardIcons: Record<string, string> = {
  amex: "fa-cc-amex",
  "diners-club": "fa-cc-diners-club",
  discover: "fa-cc-discover",
  jcb: "fa-cc-jcb",
  mastercard: "fa-cc-mastercard",
  visa: "fa-cc-visa",
};

const cardIconClass = computed(() => cardIcons[iconName] || "");

const isDefaultIcon = computed(() => !cardIconClass.value);

const iconString = computed(() => {
  if (isDefaultIcon.value) { return "fa:far fa-credit-card"; }
  return `fa:fab ${cardIconClass.value}`;
});
</script>
