<template>
  <i
    v-if="isDefaultIcon()"
    class="fa-regular fa-credit-card"
    :style="{ fontSize: size}"
    data-test="default-icon"
  />

  <i
    v-if="!isDefaultIcon()"
    :class="'fab ' + icon()"
    :style="{ fontSize: size}"
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
      amex: "fa-cc-amex",
      dinersClub: "fa-cc-diners-club",
      discover: "fa-cc-discover",
      jcb: "fa-cc-jcb",
      mastercard: "fa-cc-mastercard",
      visa: "fa-cc-visa",
    });

    const size = ref("1.5rem");

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
      size,
      isDefaultIcon,
      icon,
      convertIconName,
    };
  },
});
</script>
