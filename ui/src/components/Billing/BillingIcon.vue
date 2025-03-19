<template>
  <v-icon
    v-if="isDefaultIcon()"
    :icon="['fa:far', 'fa-credit-card'].join(' ')"
    :style="{ fontSize: size }"
    data-test="default-icon"
  />

  <v-icon
    v-else
    :icon="['fa:fab', getIcon()].join(' ')"
    :style="{ fontSize: size }"
    data-test="type-icon"
  />

</template>

<script setup lang="ts">
import { ref } from "vue";

const { iconName } = defineProps({
  iconName: {
    type: String,
    required: true,
  },
});

const cardIcon = ref({
  amex: "fa-cc-amex",
  dinersClub: "fa-cc-diners-club",
  discover: "fa-cc-discover",
  jcb: "fa-cc-jcb",
  mastercard: "fa-cc-mastercard",
  visa: "fa-cc-visa",
});

const size = ref("1.5rem");

const getFormattedIconName = () => iconName === "diners-club" ? "dinersClub" : iconName;

const isDefaultIcon = () => cardIcon.value[getFormattedIconName()] === undefined;

const getIcon = () => cardIcon.value[getFormattedIconName()] || "credit-card";
</script>
