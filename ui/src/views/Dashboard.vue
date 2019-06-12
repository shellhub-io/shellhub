<template>
<div>
  <p>Connected Devices: {{ stats.connected_devices }}</p>
  <p>Registered Devices: {{ registered_devices }}</p>
  <p>Active Sessions: {{ active_sessions }}</p>
</div>
</template>

<script>
export default {
  created() {
    this.$store.dispatch("stats/get");
  },

  computed: {
    stats() {
      return this.$store.getters["stats/stats"];
    },

    registered_devices() {
      return 0;
    },

    active_sessions() {
      return 0;
    }
  }
};
</script>
