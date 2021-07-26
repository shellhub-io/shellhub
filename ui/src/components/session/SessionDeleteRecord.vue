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
            mdi-playlist-remove
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="isOwner"
          data-test="text-tooltip"
        >
          Delete session record
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
      <v-card data-test="sessionDeleteRecord-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are going to delete the logs recorded for this session.
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
            data-test="delete-btn"
            @click="deleteRecord()"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'SessionDeleteRecord',

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

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
  },

  methods: {
    async deleteRecord() {
      try {
        await this.$store.dispatch('sessions/deleteSessionLogs', this.uid);
        this.dialog = false;
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.sessionRemoveRecord);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.sessionRemoveRecord);
      }
    },
  },
};

</script>
