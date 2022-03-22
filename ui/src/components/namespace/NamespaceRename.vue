<template>
  <fragment>
    <ValidationObserver
      ref="obs"
      v-slot="{ passes }"
    >
      <v-row>
        <v-col>
          <h3>
            Namespace
          </h3>
        </v-col>

        <v-spacer />

        <v-col
          md="auto"
          class="ml-auto"
        >
          <v-tooltip
            bottom
            :disabled="hasAuthorizationRenameNamespace"
          >
            <template #activator="{ on }">
              <div v-on="on">
                <v-btn
                  :disabled="!hasAuthorizationRenameNamespace"
                  color="primary"
                  @click="passes(editNamespace)"
                >
                  Rename Namespace
                </v-btn>
              </div>
            </template>

            <span>
              You don't have this kind of authorization.
            </span>
          </v-tooltip>
        </v-col>
      </v-row>

      <div class="mt-4 mb-2">
        <ValidationProvider
          v-slot="{ errors }"
          ref="providerName"
          vid="name"
          name="Priority"
          rules="required|rfc1123|noDot|namespace"
        >
          <v-text-field
            v-model="name"
            class="ml-3"
            label="Name"
            :error-messages="errors"
            required
            data-test="name-text"
          />
        </ValidationProvider>
      </div>
    </ValidationObserver>
  </fragment>
</template>

<script>

import {
  ValidationObserver,
  ValidationProvider,
} from 'vee-validate';

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NamespaceAddComponent',

  filters: { hasPermission },

  components: {
    ValidationProvider,
    ValidationObserver,
  },

  data() {
    return {
      name: '',
    };
  },

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    hasAuthorizationRenameNamespace() {
      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace.rename,
        );
      }

      return false;
    },
  },

  watch: {
    namespace(ns) {
      this.name = ns.name;
    },
  },

  methods: {
    async editNamespace() {
      try {
        await this.$store.dispatch('namespaces/put', { id: this.tenant, name: this.name });
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceEdit);
      } catch (error) {
        if (error.response.status === 400) {
          this.$refs.obs.setErrors({
            namespace: this.$errors.form.invalid('namespace', 'nonStandardCharacters'),
          });
        } else if (error.response.status === 409) {
          this.$refs.obs.setErrors({
            namespace: this.$errors.form.invalid('namespace', 'nameUsed'),
          });
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceEdit);
        }
      }
    },
  },
};

</script>
