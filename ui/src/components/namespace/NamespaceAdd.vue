<template>
  <fragment>
    <v-btn
      small
      class="v-btn--active"
      text
      color="primary"
      outlined
      @click="dialog = !dialog"
    >
      Add Namespace
    </v-btn>
    <v-dialog
      v-model="dialog"
      max-width="450"
      @click:outside="cancel"
    >
      <v-card>
        <v-card-title class="headline grey lighten-2 text-center">
          Enter Namespace
        </v-card-title>
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-text class="caption mb-0">
            <ValidationProvider
              v-slot="{ errors }"
              name="namespace"
              rules="required|rfc1123"
            >
              <v-text-field
                v-model="namespace"
                label="Namespace"
                :error-messages="errors"
                require
              />
            </ValidationProvider>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn
              text
              @click="cancel"
            >
              Close
            </v-btn>
            <v-btn
              color="primary"
              text
              @click="passes(addNamespace)"
            >
              Add
            </v-btn>
          </v-card-actions>
        </ValidationObserver>
      </v-card>
    </v-dialog>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

export default {
  name: 'NamespaceAdd',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      dialog: false,
      namespace: '',
    };
  },

  methods: {
    cancel() {
      this.dialog = false;
      this.$refs.obs.reset();
      this.namespace = '';
    },

    async addNamespace() {
      try {
        await this.$store.dispatch('namespaces/post', {
          name: this.namespace,
        });
        await this.$store.dispatch('namespaces/fetch');
        this.dialog = false;
        this.namespace = '';
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceCreating);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.namespaceCreating);
      }
    },
  },
};

</script>
