<template>
  <fragment>
    <v-chip
      class="float-right"
      data-test="devices-field"
    >
      {{ namespace.devices_count }}

      <font-awesome-icon
        class="ml-3 fa-lg"
        icon="microchip"
      />
    </v-chip>
  </fragment>
</template>

<script>

export default {
  name: 'MaxDevice',

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },
  },
};

</script>
