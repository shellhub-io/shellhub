<template>
  <fragment>
    <v-tooltip
      v-if="createKey"
      bottom
      :disabled="isOwner || action == 'private'"
    >
      <template #activator="{ on }">
        <div v-on="on">
          <v-btn
            :disabled="!isOwner && action == 'public'"
            class="v-btn--active"
            text
            color="primary"
            @click="dialog = !dialog"
          >
            Add {{ action }} Key
          </v-btn>
        </div>
      </template>

      <span>
        You are not the owner of this namespace
      </span>
    </v-tooltip>

    <v-tooltip
      v-else
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!isOwner && action == 'public'"
            v-on="on"
            @click="dialog = !dialog"
          >
            edit
          </v-icon>
        </span>
      </template>

      <div>
        <span v-if="action == 'private'">
          Edit
        </span>

        <span v-else>
          You are not the owner of this namespace
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card>
        <ValidationObserver
          ref="pass"
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
              vid="name"
            >
              <v-text-field
                v-model="keyLocal.name"
                label="Name"
                :error-messages="errors"
                required
              />
            </ValidationProvider>

            <ValidationProvider
              v-if="action == 'public'"
              v-slot="{ errors }"
              name="Hostname"
            >
              <v-text-field
                v-model="keyLocal.hostname"
                label="Hostname"
                :error-messages="errors"
              />
            </ValidationProvider>

            <ValidationProvider
              v-slot="{ errors }"
              vid="key"
              name="Data"
              :rules="`required|parseKey:${action}`"
              :disabled="!createKey"
            >
              <v-textarea
                v-model="keyLocal.data"
                label="Data"
                :error-messages="errors"
                required
                :disabled="!createKey"
                :messages="supportedKeys"
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
      supportedKeys: 'Supports RSA, DSA, ECDSA (nistp-*) and ED25519 key types, in PEM (PKCS#1, PKCS#8) and OpenSSH formats.',
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },
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

        if (this.action === 'public') this.keyLocal.hostname = '';
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
        } catch (error) {
          if (error.response.status === 409) {
            this.$refs.pass.setErrors({
              key: error.response.data.message,
            });
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.publicKeyCreating);
          }
        }
        break;
      case 'private':
        try {
          await this.$store.dispatch('privatekeys/set', keySend);
          this.$store.dispatch('snackbar/showSnackbarSuccessNotRequest', this.$success.privateKeyCreating);
          this.update();
        } catch (error) {
          switch (true) {
          case error.message === 'both': {
            this.$refs.pass.setErrors({
              name: ['The name already exists'],
              key: ['The private key data already exists'],
            });
            break;
          }
          case error.message === 'name': {
            this.$refs.pass.setErrors({
              name: ['The name already exists'],
            });
            break;
          }
          case error.message === 'private_key': {
            this.$refs.pass.setErrors({
              key: ['The private key data already exists'],
            });
            break;
          }
          default: {
            this.$store.dispatch('snackbar/showSnackbarErrorNotRequest', this.$errors.privateKeyCreating);
          }
          }
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
          this.$store.dispatch('snackbar/showSnackbarSuccessNotRequest', this.$success.privateKeyEditing);
          this.update();
        } catch (error) {
          if (error.message === 'name') {
            this.$refs.pass.setErrors({
              name: ['The name already exists'],
            });
          } else {
            this.$store.dispatch('snackbar/showSnackbarErrorNotRequest', this.$errors.privateKeyEditing);
          }
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
