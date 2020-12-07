<template>
  <fragment>
    <v-btn
      color="red darken-1"
      outlined
      @click="dialog = !dialog"
    >
      Delete namespace
    </v-btn>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          Once you delete a namespace, there is no going back. Please be certain.
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
  name: 'NamespaceDelete',

  props: {
    nsTenant: {
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
    tenant() {
      return this.$props.nsTenant;
    },
  },

  methods: {
    async remove() {
      try {
        await this.$store.dispatch('namespaces/remove', this.tenant);
        this.dialog = !this.dialog;
        this.$store.dispatch('auth/logout').then(() => {
          this.$router.push('/login');
        });
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceDelete);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceDelete);
      }
    },
  },
};

</script>
