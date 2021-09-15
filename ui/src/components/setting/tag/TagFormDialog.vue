<template>
  <fragment>
    <v-tooltip
      v-if="isCreate"
      bottom
    >
      <template #activator="{ on }">
        <v-icon
          :disabled="!isOwner"
          v-on="on"
          @click="dialog = !dialog"
        >
          mdi-tag
        </v-icon>
      </template>

      <div>
        <span
          v-if="isOwner"
          data-test="text-span"
        >
          Create tag
        </span>

        <span v-else>
          You are not the owner of this namespace
        </span>
      </div>
    </v-tooltip>

    <v-tooltip
      v-else
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-icon
            :disabled="!isOwner"
            v-on="on"
            @click="dialog = !dialog"
          >
            edit
          </v-icon>
        </span>
      </template>

      <div>
        <span
          v-if="isOwner"
          data-test="text-span"
        >
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
      <v-card data-test="tagForm-card">
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title class="headline grey lighten-2 text-center">
            <div v-if="isCreate">
              New tag
            </div>
            <div v-else>
              Edit tag
            </div>
          </v-card-title>

          <v-card-text>
            <ValidationProvider
              v-slot="{ errors }"
              ref="providerTag"
              name="Name"
              rules="required|tag|routeIdentifier"
            >
              <v-text-field
                v-model="tagLocal"
                label="Name"
                :error-messages="errors"
                required
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="cancel-btn"
              @click="close"
            >
              Cancel
            </v-btn>

            <v-btn
              text
              data-test="doAction-btn"
              @click="passes(doAction)"
            >
              {{ action }}
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
  name: 'TagFormDialog',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
    action: {
      type: String,
      default: 'create',
      required: false,
      validator: (value) => ['create', 'edit'].includes(value),
    },

    uid: {
      type: String,
      default: '',
      required: false,
    },

    tagName: {
      type: String,
      default: '',
      required: false,
    },
  },

  data() {
    return {
      dialog: false,
      tagLocal: String,
    };
  },

  computed: {
    isOwner() {
      return this.$store.getters['namespaces/owner'];
    },

    isCreate() {
      return this.action === 'create';
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
      if (this.isCreate) {
        this.tagLocal = '';
      } else {
        this.tagLocal = this.tagName;
      }
    },

    async doAction() {
      if (this.isCreate) {
        try {
          await this.$store.dispatch('tags/post', { uid: this.uid, name: this.tagLocal });

          this.update();
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceTagCreate);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagCreate);
        }
      } else {
        try {
          await this.$store.dispatch('tags/edit', { oldTag: this.tagName, newTag: this.tagLocal });

          this.update();
          this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceTagEdit);
        } catch {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagEdit);
        }
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
