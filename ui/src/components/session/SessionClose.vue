<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          mdi-close-circle
        </v-icon>
      </template>
      <span>Close</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are going to close connection for this device
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            color="red darken-1"
            text
            @click="dialog=!dialog"
          >
            Cancel
          </v-btn>

          <v-btn
            text
            @click="close();"
          >
            Close
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'SessionClose',

  props: {
    session: {
      type: Object,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods: {
    async close() {
      await this.$store.dispatch('sessions/close', this.session);
      this.dialog = false;
      this.$emit('update');
    },
  },
};

</script>
