<template>
  <fragment>
    <v-app-bar
      v-if="isLoggedIn"
      app
      flat
    >
      <v-app-bar-nav-icon class="hidden-lg-and-up" />

      <router-link to="/" />

      <v-spacer />

      <v-menu
        offset-y
      >
        <template #activator="{ on }">
          <v-btn
            color="primary"
            text
            v-on="on"
          >
            <v-icon
              :size="defaultSize"
              class="ml-1"
              left
            >
              mdi-account
            </v-icon>
            <div
              v-if="!isMobile"
              class="ml-1 mr-1"
            >
              {{ $store.getters["auth/currentUser"] }}
            </div>
            <v-icon
              :size="defaultSize"
              class="ml-1 mr-1"
              right
            >
              mdi-chevron-down
            </v-icon>
          </v-btn>
        </template>

        <v-card>
          <v-list-item
            v-for="(item, index) in menu"
            :key="index"
            router
            :data-test="item.title"
            @click.prevent="triggerClick(item)"
          >
            <v-icon left>
              {{ item.icon }}
            </v-icon>
            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </v-list-item>
        </v-card>
      </v-menu>
    </v-app-bar>

    <aside
      ref="chat"
      class="gitter-chat-embed is-collapsed"
    />
  </fragment>
</template>

<script>

import GitterSidecar from 'gitter-sidecar';

export default {
  name: 'AppBarComponent',

  data() {
    return {
      clipped: false,
      chat: null,
      chatOpen: false,
      defaultSize: 24,
      menu: [
        {
          title: 'Settings',
          type: 'path',
          path: '/settings',
          icon: 'mdi-cog',
          items: [{ title: 'Profile', path: '/settings/profile' }],
        },
        {
          title: 'Logout',
          type: 'method',
          icon: 'mdi-logout',
          method: 'logout',
        },
      ],
    };
  },

  computed: {
    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },

    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    isMobile() {
      return this.$store.getters['mobile/isMobile'];
    },
  },

  async mounted() {
    this.chat = await new GitterSidecar({ room: 'shellhub-io/community', activationElement: false, targetElement: this.$refs.chat });
    this.$refs.chat.addEventListener('gitter-chat-toggle', (e) => {
      this.chatOpen = e.detail.state;
    });
  },

  methods: {
    async logout() {
      try {
        this.$store.dispatch('auth/logout');
        await this.$router.push('/login');

        this.$store.dispatch('layout/setLayout', 'simpleLayout');
        this.$store.dispatch('namespaces/clearNamespaceList');
      } catch {
        this.$store.dispatch('snackbar/showSnackbarErrorNotRequest', this.$errors.snackbar.logoutFailed);
      }
    },

    triggerClick(item) {
      switch (item.type) {
      case 'path':
        this.$router.push(item.path).catch(() => {});
        break;
      case 'method':
        this[item.method]();
        break;
      default:
        break;
      }
    },

    toggleChat() {
      this.chat.toggleChat(!this.chatOpen);
    },
  },
};

</script>
