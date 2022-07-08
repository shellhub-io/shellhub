<template>
  <font-awesome-icon
    v-if="isDefaultIcon()"
    icon="credit-card"
    size="lg"
    data-test="default-icon"
  />

  <font-awesome-icon
    v-if="!isDefaultIcon()"
    :icon="['fab', icon()]"
    size="lg"
    data-test="type-icon"
  />
</template>

<script lang="ts">
import { defineComponent, ref } from "vue";

export default defineComponent({
  props: {
    iconName: {
      type: String,
      required: true,
    },
  },
  setup(props) {
    const cardIcon = ref({
      amex: "cc-amex",
      dinersClub: "cc-diners-club",
      discover: "cc-discover",
      jcb: "cc-jcb",
      mastercard: "cc-mastercard",
      visa: "cc-visa",
    });

    const isDefaultIcon = () => {
      // @ts-ignore
      return cardIcon.value[convertIconName()] === undefined;
    };

    const icon = () => {
      // @ts-ignore
      return cardIcon.value[convertIconName()] || "credit-card";
    };

    const convertIconName = () => {
      if (props.iconName === "diners-club") {
        return "dinersClub";
      }

      return props.iconName;
    };

    return {
      cardIcon,
      isDefaultIcon,
      icon,
      convertIconName,
    };
  },
});
</script>
