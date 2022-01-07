<template>
  <fragment>
    <v-tooltip
      v-if="isCreate"
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title
            data-test="create-item"
            v-on="on"
          >
            Add tag
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="create-icon"
            v-on="on"
          >
            mdi-tag
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-tooltip
      v-else
      :disabled="hasAuthorization"
      bottom
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title
            data-test="edit-item"
            v-on="on"
          >
            Edit
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="edit-icon"
            v-on="on"
          >
            mdi-tag
          </v-icon>
        </span>
      </template>

      <span v-if="!hasAuthorization">
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="tagForm-card">
        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
          <v-card-title class="headline primary text-center">
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'TagFormDialogComponent',

  filters: { hasPermission },

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

    show: {
      type: Boolean,
      required: true,
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

    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },
      set(value) {
        this.$emit('update:show', value);
      },
    },

    hasAuthorization() {
      const role = this.$store.getters['auth/role'];

      if (role !== '') {
        let action = '';
        if (this.isCreate) action = 'deviceCreate';
        else action = 'edit';

        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.tag[action],
        );
      }

      return false;
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
      this.$emit('update:show', false);
      this.$refs.obs.reset();
    },
  },
};
</script>
