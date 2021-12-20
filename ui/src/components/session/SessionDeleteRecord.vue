<template>
  <fragment>
    <v-tooltip
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title data-test="play-item">
            Delete Session Record
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="play-icon"
            v-on="on"
          >
            mdi-playlist-remove
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
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
            @click="close()"
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'SessionDeleteRecordComponent',

  filters: { hasPermission },

  props: {
    uid: {
      type: String,
      required: true,
    },

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      action: 'removeRecord',
    };
  },

  computed: {
    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },
      set(value) {
        this.$emit('update:show', value);
      },
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.session[this.action],
        );
      }

      return false;
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

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
