<template>
  <fragment>
    <v-tooltip
      bottom
      :disabled="hasAuthorization"
    >
      <template #activator="{ on }">
        <span v-on="on">
          <v-btn
            :disabled="!hasAuthorization"
            class="mr-2"
            outlined
            data-test="removeMember-btn"
            @click="dialog = !dialog"
          >
            <v-icon
              outlined
              :disabled="!hasAuthorization"
              v-on="on"
            >
              delete
            </v-icon>
          </v-btn>
        </span>
      </template>

      <div>
        <span
          v-if="hasAuthorization"
          data-test="tooltip-text"
        >
          Remove
        </span>

        <span
          v-else
        >
          You don't have this kind of authorization.
        </span>
      </div>
    </v-tooltip>

    <v-dialog
      v-model="dialog"
      max-width="400"
    >
      <v-card data-test="namespaceMemberDelete-dialog">
        <v-card-title class="headline grey lighten-2 text-center">
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
            @click="dialog=!dialog"
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
  },

  data() {
    return {
      dialog: false,
    };
  },

  computed: {
    hasAuthorization() {
      const ownerID = this.$store.getters['namespaces/get'].owner;
      if (this.member.id === ownerID) {
        return false;
      }

      const accessType = this.$store.getters['auth/accessType'];
      if (accessType !== '') {
        let action = '';
        if (this.addUser) action = 'addMember';
        else action = 'removeMember';

        return hasPermission(
          this.$authorizer.accessType[accessType],
          this.$actions.namespace[action],
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
      this.dialog = false;
    },
  },
};

</script>
