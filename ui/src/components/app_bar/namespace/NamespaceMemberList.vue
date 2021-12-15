<template>
  <fragment>
    <div class="mt-5">
      <v-data-table
        class="elevation-0"
        :headers="headers"
        :items="members"
        hide-default-footer
        data-test="dataTable-field"
      >
        <template #[`item.username`]="{ item }">
          <v-icon>
            mdi-account
          </v-icon>
          {{ item.username }}
        </template>

        <template #[`item.role`]="{ item }">
          {{ item.role }}
        </template>

        <template #[`item.actions`]="{ item }">
          <v-menu
            v-if="item.id !== namespace.owner"
            :ref="'menu'+members.indexOf(item)"
            offset-y
          >
            <template #activator="{ on, attrs}">
              <v-icon
                small
                class="icons"
                v-bind="attrs"
                v-on="on"
              >
                mdi-dots-horizontal
              </v-icon>
            </template>

            <v-list>
              <v-list-item @click.stop="showNamespaceMemberForm(members.indexOf(item))">
                <NamespaceMemberFormDialog
                  :add-user="false"
                  :member="item"
                  :show.sync="namespaceMemberFormShow[members.indexOf(item)]"
                  data-test="NamespaceMemberFormDialogEdit-component"
                  @update="refresh"
                />
              </v-list-item>

              <v-list-item @click.stop="showNamespaceMemberDelete(members.indexOf(item))">
                <NamespaceMemberDelete
                  :member="item"
                  :show.sync="namespaceMemberDeleteShow[members.indexOf(item)]"
                  data-test="namespaceMemberDelete-component"
                  @update="refresh"
                />
              </v-list-item>
            </v-list>
          </v-menu>
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
      menu: false,
      namespaceMemberFormShow: [],
      namespaceMemberDeleteShow: [],
      headers: [
        {
          text: 'Username',
          value: 'username',
          align: 'start',
          sortable: false,
        },
        {
          text: 'Role',
          value: 'role',
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

    members() {
      return this.namespace.members;
    },
  },

  methods: {
    refresh() {
      this.getNamespace();
    },

    showNamespaceMemberForm(index) {
      this.namespaceMemberFormShow[index] = this.namespaceMemberFormShow[index] === undefined
        ? true : !this.namespaceMemberFormShow[index];
      this.$set(this.namespaceMemberFormShow, index, this.namespaceMemberFormShow[index]);

      this.closeMenu(index);
    },

    showNamespaceMemberDelete(index) {
      this.namespaceMemberDeleteShow[index] = this.namespaceMemberDeleteShow[index] === undefined
        ? true : !this.namespaceMemberDeleteShow[index];
      this.$set(this.namespaceMemberDeleteShow, index, this.namespaceMemberDeleteShow[index]);

      this.closeMenu(index);
    },

    closeMenu(index) {
      this.$refs[`menu${index}`].isActive = false;
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
