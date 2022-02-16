<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="remove-icon"
        v-text="'delete'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="remove-title"
        v-text="'Remove'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="publicKeyDelete-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Are you sure?'"
        />

        <v-card-text
          class="mt-4 mb-3 pb-1"
          data-test="text-text"
          v-text="'You are about to remove this public key.'"
        />

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="close-btn"
            @click="close()"
            v-text="'Close'"
          />

          <v-btn
            color="red darken-1"
            text
            data-test="remove-btn"
            @click="remove()"
            v-text="'Remove'"
          />
        </v-card-actions>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

export default {
  name: 'PublicKeyDeleteComponent',

  props: {
    fingerprint: {
      type: String,
      required: true,
    },

    show: {
      type: Boolean,
      required: false,
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
    async remove() {
      try {
        await this.$store.dispatch('publickeys/remove', this.fingerprint);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyDeleting);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.publicKeyDeleting);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};
</script>
