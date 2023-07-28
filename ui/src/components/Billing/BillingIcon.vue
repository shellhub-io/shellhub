<template>
  <v-icon
    v-if="isDefaultIcon()"
    :icon="['fa:far', 'fa-credit-card'].join(' ')"
    :style="{ fontSize: size }"
    data-test="default-icon"
  />

  <v-icon
    v-else
    :icon="['fa:fab', icon()].join(' ')"
    :style="{ fontSize: size }"
    data-test="type-icon"
  />

</template>

<script lang="ts">
/* eslint-disable */
import { defineComponent, ref } from "vue";
import { library } from '@fortawesome/fontawesome-svg-core';
import { faCreditCard } from '@fortawesome/free-regular-svg-icons';
import { faCcAmex, faCcDinersClub, faCcDiscover, faCcJcb, faCcMastercard, faCcVisa } from '@fortawesome/free-brands-svg-icons';

library.add(
  faCreditCard,
  faCcAmex,
  faCcDinersClub,
  faCcDiscover,
  faCcJcb,
  faCcMastercard,
  faCcVisa
);

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

    const isDefaultIcon = () =>
      cardIcon.value[convertIconName()] === undefined;
    const icon = () =>
      cardIcon.value[convertIconName()] || "credit-card";
    const convertIconName = () => {
      props.iconName === "diners-club" ? "dinersClub" : props.iconName

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
