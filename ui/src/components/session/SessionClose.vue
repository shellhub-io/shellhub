<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="close-icon"
        v-text="'mdi-close-circle'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="close-title"
        v-text="'Close'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="sessionClose-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Are you sure?'"
        />

        <v-card-text
          class="mt-4 mb-3 pb-1"
          data-test="text-text"
          v-text="'You are going to close connection for this device.'"
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
            data-test="close-btn"
            @click="closeSession()"
            v-text="'Close'"
          />
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

    show: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      session: {},
    };
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

  created() {
    this.session = {
      uid: this.uid,
      device_uid: this.device,
    };
  },

  methods: {
    async closeSession() {
      try {
        await this.$store.dispatch('sessions/close', this.session);
        this.close();

        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.sessionClose);
        this.$emit('update');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.sessionClose);
      }
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
