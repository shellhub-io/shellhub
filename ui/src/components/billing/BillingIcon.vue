<template>
  <fragment>
    <font-awesome-icon
      v-if="isDefaultIcon()"
      class="ml-2"
      icon="credit-card"
      size="lg"
      data-test="default-icon"
    />

    <font-awesome-icon
      v-if="!isDefaultIcon()"
      class="ml-2"
      :icon="[ 'fab', icon() ]"
      size="lg"
      data-test="type-icon"
    />
  </fragment>
</template>

<script>

export default {
  name: 'BillingIconComponent',

  props: {
    iconName: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      cardIcon: {
        amex: 'cc-amex',
        dinersClub: 'cc-diners-club',
        discover: 'cc-discover',
        jcb: 'cc-jcb',
        mastercard: 'cc-mastercard',
        visa: 'cc-visa',
      },
    };
  },

  methods: {
    isDefaultIcon() {
      return this.cardIcon[this.convertIconName()] === undefined;
    },

    icon() {
      return this.cardIcon[this.convertIconName()] || 'credit-card';
    },

    convertIconName() {
      if (this.iconName === 'diners-club') {
        return 'dinersClub';
      }

      return this.iconName;
    },
  },
};

</script>
