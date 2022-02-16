<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="edit-icon"
        v-text="'mdi-tag'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="edit-title"
        v-text="'Edit'"
      />
    </v-list-item-content>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="tagForm-card">
        <v-card-title
          class="headline primary"
          data-test="text-title"
          v-text="'Edit tag'"
        />

        <ValidationObserver
          ref="obs"
          v-slot="{ passes }"
        >
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
                data-test="name-field"
              />
            </ValidationProvider>
          </v-card-text>

          <v-card-actions>
            <v-spacer />

            <v-btn
              text
              data-test="cancel-btn"
              @click="close"
              v-text="'Cancel'"
            />

            <v-btn
              text
              data-test="edit-btn"
              @click="passes(edit)"
              v-text="'Edit'"
            />
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
  name: 'TagFormDialogComponent',

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  props: {
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
      tagLocal: '',
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

  async updated() {
    await this.setLocalVariable();
  },

  methods: {
    setLocalVariable() {
      this.tagLocal = this.tagName;
    },

    async edit() {
      try {
        await this.$store.dispatch('tags/edit', { oldTag: this.tagName, newTag: this.tagLocal });

        this.update();
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.deviceTagEdit);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.deviceTagEdit);
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
