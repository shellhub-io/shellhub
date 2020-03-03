<template>
  <v-dialog v-model="show" max-width="800px">
    <v-card>
      
      <v-card-title>Run the following command to register your device:</v-card-title>
      <v-card-text>
        <code class="pa-2">
          curl "http://{{ hostname }}/install.sh?tenant_id={{ tenant }}" | sh
        </code>
      </v-card-text>
      
      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn text @click="show = false">Close</v-btn>
      </v-card-actions>
    
    </v-card>
</v-dialog>
</template>

<script>
export default {
  name: "DeviceAdd",

  data() {
    return {
      hostname: window.location.hostname
    };
  },

  computed: {
    tenant() {
      return this.$store.getters["auth/tenant"];
    },

    show: {
      get() {
        return this.$store.getters["modals/add_device"];
      },

      set(value) {
        this.$store.dispatch("modals/showAddDevice", value);
      }
    }
  }
};
</script>
