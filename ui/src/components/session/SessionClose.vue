<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!isOwner"
            v-on="on"
            @click="dialog = !dialog"
          >
            mdi-close-circle
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="isOwner"
          data-test="text-tooltip"
        >
          Close
        </span>

        <span v-else>
          You are not the owner of this namespace
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="sessionClose-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are going to close connection for this device.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="cancel-btn"
            @click="dialog=!dialog"
          >
            Cancel
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            data-test="close-btn"
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
  name: 'SessionCloseComponent',

  props: {
    uid: {
      type: String,
      required: true,
    },
    device: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      session: {},
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  created() {
    this.session = {
      uid: this.uid,
      device_uid: this.device,
    };
  },

  methods: {
    async close() {
      try {
        await this.$store.dispatch('sessions/close', this.session);
        this.dialog = false;

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.sessionClose);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.sessionClose);
      }
    },
  },
};

</script>
