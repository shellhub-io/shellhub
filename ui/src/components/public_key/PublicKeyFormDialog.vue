<template>
  <fragment>
    <v-btn
      v-if="createPublicKey"
      class="v-btn--active"
      text
      color="primary"
      @click="dialog = !dialog"
    >
      Add Public Key
    </v-btn>
    <v-tooltip
      v-else
      bottom
    >
      <template #activator="{ on }">
        <v-icon
          v-on="on"
          @click="dialog = !dialog"
        >
          edit
        </v-icon>
      </template>
      <span>Edit</span>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card>
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title
            v-if="createPublicKey"
            class="headline grey lighten-2 text-center"
          >
            New Public Key
          </v-card-title>
          <v-card-title
            v-else
            class="headline grey lighten-2 text-center"
          >
            Edit Public Key
          </v-card-title>

          <v-card-text>
            <ValidationProvider
              v-slot="{ errors }"
              name="Name"
              rules="required"
            >
              <v-text-field
                v-model="publicKeyLocal.name"
                label="Name"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Data"
              rules="required"
              :disabled="!createPublicKey"
            >
              <v-textarea
                v-model="publicKeyLocal.data"
                label="Data"
                :error-messages="errors"
                required
                :disabled="!createPublicKey"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              @click="close"
            >
              Cancel
            </v-btn>

            <v-btn
              v-if="createPublicKey"
              text
              @click="passes(create)"
            >
              Create
            </v-btn>

            <v-btn
              v-else
              text
              @click="passes(edit)"
            >
              Edit
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
  name: 'PublicKeyFormDialog',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    publicKey: {
      type: Object,
      required: false,
      default: Object,
    },

    createPublicKey: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
      dialog: false,
      publicKeyLocal: [],
    };
  },

  async created() {
    await this.setLocalVariable();
  },

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      if (this.createPublicKey) {
        this.publicKeyLocal = {
          name: '',
          data: '',
        };
      } else {
        this.publicKeyLocal = { ...this.publicKey };
        this.publicKeyLocal.data = atob(this.publicKey.data);
      }
    },

    async create() {
      const publicKeySend = this.publicKeyLocal;
      publicKeySend.data = btoa(this.publicKeyLocal.data);

      try {
        await this.$store.dispatch('publickeys/post', publicKeySend);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyCreating);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyCreating);
      }
    },

    async edit() {
      try {
        await this.$store.dispatch('publickeys/put', this.publicKeyLocal);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyEditing);
        this.update();
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyEditing);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = !this.dialog;
      this.$refs.obs.reset();
    },
  },
};
</script>
