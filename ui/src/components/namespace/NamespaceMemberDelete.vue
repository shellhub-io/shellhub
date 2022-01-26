<template>
  <fragment>
    <v-tooltip
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-list-item-title data-test="remove-item">
            Remove
          </v-list-item-title>
        </span>

        <span v-on="on">
          <v-icon
            :disabled="!hasAuthorization"
            left
            data-test="remove-icon"
            v-on="on"
          >
            delete
          </v-icon>
        </span>
      </template>

      <span>
        You don't have this kind of authorization.
      </span>
    </v-tooltip>

    <v-dialog
      v-model="showDialog"
      max-width="400"
      @click:outside="close"
    >
      <v-card data-test="namespaceMemberDelete-dialog">
        <v-card-title class="headline primary">
          Are you sure?
        </v-card-title>

        <v-card-text class="mt-4 mb-3 pb-1">
          You are about to remove this user from the namespace.
        </v-card-text>

        <v-card-actions>
          <v-spacer />

          <v-btn
            text
            data-test="close-btn"
            @click="close()"
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

import hasPermission from '@/components/filter/permission';

export default {
  name: 'NamespaceNewMemberComponent',

  filters: { hasPermission },

  props: {
    member: {
      type: Object,
      required: true,
    },

    show: {
      type: Boolean,
      required: false,
      default: false,
    },
  },

  data() {
    return {
      action: 'removeMember',
    };
  },

  computed: {
    showDialog: {
      get() {
        return this.show && this.hasAuthorization;
      },

      set(value) {
        this.$emit('update:show', value);
      },
    },

    hasAuthorization() {
      const ownerID = this.$store.getters['namespaces/get'].owner;
      if (this.member.id === ownerID) {
        return false;
      }

      const role = this.$store.getters['auth/role'];
      if (role !== '') {
        return hasPermission(
          this.$authorizer.role[role],
          this.$actions.namespace[this.action],
        );
      }

      return false;
    },
  },

  methods: {
    async remove() {
      try {
        const tenant = this.$store.getters['auth/tenant'];
        await this.$store.dispatch('namespaces/removeUser', {
          user_id: this.member.id,
          tenant_id: tenant,
        });

        this.update();
        this.$store.dispatch('snackbar/showSnackbarSuccessAction', this.$success.namespaceRemoveUser);
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceRemoveUser);
      }
    },

    update() {
      this.$emit('update');
      this.close();
    },

    close() {
      this.$emit('update:show', false);
    },
  },
};

</script>
