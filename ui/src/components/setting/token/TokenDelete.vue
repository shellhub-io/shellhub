<template>
  <fragment>
    <v-tooltip bottom>
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          delete
        </v-icon>
      </template>
      <span>
        Remove
      </span>
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
          You are about to remove this token.
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
            color="red darken-1"
            text
            @click="remove();"
          >
            Remove
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'DeviceAdd',

  props: {
    id: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      token: {},
    };
  },

  methods: {
    async remove() {
      try {
        await this.$store.dispatch('tokens/remove', this.id);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.tokenDeleting);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.tokenDeleting);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = !this.dialog;
    },
  },
};

</script>
