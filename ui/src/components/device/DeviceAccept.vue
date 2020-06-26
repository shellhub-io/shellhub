<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          check
        </v-icon>
      </template>
      <span>Check</span>
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
          You are about to accept this device
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            text
            @click="accept();"
          >
            Accept
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'DeviceAccept',

  props: {
    uid: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
    };
  },

  methods: {
    async accept() {
      await this.$store.dispatch('devices/accept', this.uid);
      this.dialog = !this.dialog;
      this.$router.push('/devices');
    },
  },
};

</script>
