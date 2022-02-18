<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="removeRecord-icon"
        v-text="'mdi-playlist-remove'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="removeRecord-title"
        v-text="'Delete Session Record'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="sessionDeleteRecord-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Are you sure?'"
        />

        <v-card-text
          class="mt-4 mb-3 pb-1"
          data-test="text-text"
          v-text="'You are going to delete the logs recorded for this session.'"
        />

        <v-card-actions>
          <v-spacer />
          <v-btn
            text
            data-test="cancel-btn"
            @click="close()"
            v-text="'Cancel'"
          />

          <v-btn
            color="red darken-1"
            text
            data-test="delete-btn"
            @click="deleteRecord()"
            v-text="'Delete'"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'SessionDeleteRecordComponent',

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

  computed: {
    showDialog: {
      get() {
        return this.show;
      },
      set(value) {
        this.$emit('update:show', value);
      },
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
