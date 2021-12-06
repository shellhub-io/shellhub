<template>
  <fragment>
    <div class="mt-5">
      <v-row class="text-center mb-2">
        <v-col
          v-for="item in heading"
          :key="item.id"
        >
          <b :data-test="item.title+'-title'">
            {{ item.title }}
          </b>
        </v-col>
      </v-row>

      <v-list class="mb-2">
        <v-list-item
          v-for="item in namespace.members"
          :key="item.id"
        >
          <v-row>
            <v-col cols="1">
              <v-icon>
                mdi-account
              </v-icon>
            </v-col>

            <v-col class="text-start">
              <v-list-item-title :data-test="item.username+'-list'">
                {{ item.username }}
              </v-list-item-title>
            </v-col>

            <v-col
              cols="3"
              class="text-end"
            >
              <v-list-item-title :data-test="item.type+'-list'">
                {{ item.type }}
              </v-list-item-title>
            </v-col>

            <v-spacer />

            <div :data-test="item.username+'-actions-list'">
              <v-col>
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
              </v-col>
            </div>
          </v-row>
        </v-list-item>
      </v-list>
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
      heading: [
        {
          id: 'name',
          title: 'Username',
        },
        {
          id: 'role',
          title: 'Role',
        },
        {
          id: 'actions',
          title: 'Actions',
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
