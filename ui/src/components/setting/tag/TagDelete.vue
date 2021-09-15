<template>
  <fragment>
    <v-tooltip
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!isOwner"
            v-on="on"
            @click="dialog = !dialog"
          >
            delete
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="isOwner"
          data-test="text-span"
        >
          Remove
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
      <v-card data-test="tagDelete-card">
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this tag.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="close-btn"
            @click="dialog=!dialog"
          >
            Close
          </v-btn>

          <v-btn
            color="red darken-1"
            text
            data-test="remove-btn"
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
  name: 'TagDelete',

  props: {
    tagName: {
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
    async remove() {
      try {
        await this.$store.dispatch('tags/remove', this.tagName);
        this.update();
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceTagRemove);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagRemove);
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
