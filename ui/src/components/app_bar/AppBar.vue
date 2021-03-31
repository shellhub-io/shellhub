<template>
  <fragment>
    <v-app-bar
      v-if="isLoggedIn"
      app
      clipped-left
      dark
      color="primary"
    >
      <v-app-bar-nav-icon
        class="hidden-lg-and-up"
        @click.stop="drawer = !drawer"
      />

      <router-link to="/">
        <v-img
          src="@/assets/logo-inverted.png"
          max-width="160"
        />
      </router-link>

      <span class="overline mt-3">BETA</span>

      <v-spacer />

      <NamespaceMenu
        :in-a-namespace="hasNamespaces"
      />

      <v-chip>
        <v-icon
          @click="toggleChat()"
        >
          help
        </v-icon>
      </v-chip>

      <Notification
        :in-a-namespace="hasNamespaces"
      />

      <v-menu
        offset-y
      >
        <template #activator="{ on }">
          <v-chip v-on="on">
            <v-icon left>
              mdi-account
            </v-icon>
            {{ $store.getters["auth/currentUser"] }}
            <v-icon right>
              mdi-chevron-down
            </v-icon>
          </v-chip>
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
import NamespaceMenu from '@/components/app_bar/namespace/NamespaceMenu';
import Notification from '@/components/app_bar/notification/Notification';

export default {
  name: 'AppBar',

  components: {
    NamespaceMenu,
    Notification,
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      chat: null,
      chatOpen: false,
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
  },

  async mounted() {
    this.chat = await new GitterSidecar({ room: 'shellhub-io/community', activationElement: false, targetElement: this.$refs.chat });
    this.$refs.chat.addEventListener('gitter-chat-toggle', (e) => {
      this.chatOpen = e.detail.state;
    });
  },

  methods: {
    logout() {
      this.$store.dispatch('auth/logout').then(() => {
        this.$store.dispatch('namespaces/clearNamespaceList');
        this.$router.push('/login');
      });
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
