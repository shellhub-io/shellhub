<template>
  <fragment>
    <v-btn
      v-if="createKey"
      class="v-btn--active"
      text
      color="primary"
      @click="dialog = !dialog"
    >
      Add {{ action }} Key
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
            v-if="createKey"
            class="headline grey lighten-2 text-center"
          >
            New {{ action }} key
          </v-card-title>
          <v-card-title
            v-else
            class="headline grey lighten-2 text-center"
          >
            Edit {{ action }} key
          </v-card-title>

          <v-card-text>
            <ValidationProvider
              v-slot="{ errors }"
              name="Name"
              rules="required"
            >
              <v-text-field
                v-model="keyLocal.name"
                label="Name"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              name="Data"
              rules="required|parseKey"
              :disabled="!createKey"
            >
              <v-textarea
                v-model="keyLocal.data"
                label="Data"
                :error-messages="errors"
                required
                :disabled="!createKey"
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
              v-if="createKey"
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
  name: 'KeyFormDialog',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    keyObject: {
      type: Object,
      required: false,
      default: Object,
    },

    createKey: {
      type: Boolean,
      required: true,
    },

    action: {
      type: String,
      default: 'public',
      required: false,
      validator: (value) => ['public', 'private'].includes(value),
    },
  },

  data() {
    return {
      dialog: false,
      keyLocal: [],
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
      if (this.createKey) {
        this.keyLocal = {
          name: '',
          data: '',
        };
      } else {
        this.keyLocal = { ...this.keyObject };
        if (this.action === 'public') this.keyLocal.data = atob(this.keyObject.data);
      }
    },

    async create() {
      const keySend = this.keyLocal;

      switch (this.action) {
      case 'public':
        try {
          keySend.data = btoa(this.keyLocal.data);
          await this.$store.dispatch('publickeys/post', keySend);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyCreating);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyCreating);
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/set', keySend);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.privateKeyCreating);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.privateKeyCreating);
        }
        break;
      default:
      }
    },

    async edit() {
      switch (this.action) {
      case 'public':
        try {
          await this.$store.dispatch('publickeys/put', this.keyLocal);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.publicKeyEditing);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyEditing);
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/edit', this.keyLocal);
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.privateKeyEditing);
          this.update();
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.privateKeyEditing);
        }
        break;
      default:
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.dialog = !this.dialog;
      this.$refs.pass.reset();
    },
  },
};
</script>
