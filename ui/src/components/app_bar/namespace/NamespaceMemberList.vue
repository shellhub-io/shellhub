<template>
  <fragment>
    <div class="mt-5">
      <v-data-table
        class="elevation-0"
        :headers="headers"
        :items="namespace.members"
        hide-default-footer
        data-test="dataTable-field"
      >
        <template #[`item.username`]="{ item }">
          <v-icon>
            mdi-account
          </v-icon>
          {{ item.username }}
        </template>

        <template #[`item.type`]="{ item }">
          {{ item.type }}
        </template>

        <template #[`item.actions`]="{ item }">
          <NamespaceMemberFormDialog
            :add-user="false"
            :member="item"
            data-test="NamespaceMemberFormDialogEdit-component"
            @update="refresh"
          />

          <NamespaceMemberDelete
            :member="item"
            data-test="namespaceMemberDelete-component"
            @update="refresh"
          />
        </template>
      </v-data-table>
    </div>
  </fragment>
</template>

<script>

import NamespaceMemberDelete from '@/components/app_bar/namespace/NamespaceMemberDelete';
import NamespaceMemberFormDialog from '@/components/app_bar/namespace/NamespaceMemberFormDialog';

export default {
  name: 'NamespaceMemberList',

  components: {
    NamespaceMemberDelete,
    NamespaceMemberFormDialog,
  },

  props: {
    namespace: {
      type: Object,
      required: true,
    },
  },

  data() {
    return {
      headers: [
        {
          text: 'Username',
          value: 'username',
          align: 'start',
          sortable: false,
        },
        {
          text: 'Role',
          value: 'type',
          align: 'center',
          sortable: false,
        },
        {
          text: 'Actions',
          value: 'actions',
          align: 'end',
          sortable: false,
        },
      ],
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },
  },

  methods: {
    refresh() {
      this.getNamespace();
    },

    async getNamespace() {
      try {
        await this.$store.dispatch('namespaces/get', this.tenant);
        this.name = this.namespace.name;
      } catch (error) {
        if (error.response.status === 403) {
          this.$store.dispatch('snackbar/showSnackbarErrorAssociation');
        } else {
          this.$store.dispatch('snackbar/showSnackbarErrorAction', this.$errors.snackbar.namespaceLoad);
        }
      }
    },
  },
};

</script>
