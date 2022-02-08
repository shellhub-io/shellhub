<template>
  <fragment>
    <v-list-item-icon class="mr-0">
      <v-icon
        left
        data-test="remove-icon"
        v-text="'delete'"
      />
    </v-list-item-icon>

    <v-list-item-content>
      <v-list-item-title
        class="text-left"
        data-test="remove-title"
        v-text="'Remove'"
      />
    </v-list-item-content>

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
