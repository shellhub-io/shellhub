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
        @click.stop="updateDrawer()"
      />

      <router-link to="/">
        <v-img
          class="d-sm-flex hidden-sm-and-down"
          src="@/assets/logo-inverted.png"
          max-width="160"
        />

        <v-img
          class="hidden-sm-and-up"
          src="@/assets/logo-inverted-only-cloud.png"
          max-width="46"
        />
      </router-link>

      <span
        class="overline mt-3 d-sm-flex hidden-sm-and-down"
      >
        BETA
      </span>

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

      <Notification />

      <v-menu
        offset-y
      >
        <template #activator="{ on }">
          <v-chip v-on="on">
            <v-icon left>
              mdi-account
            </v-icon>
            <div
              v-if="!isMobile"
            >
              {{ $store.getters["auth/currentUser"] }}
            </div>
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

  props: {
    drawer: {
      type: Boolean,
      required: true,
    },
  },

  data() {
    return {
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

    updateDrawer() {
      this.$emit('update:drawer', !this.drawer);
    },
  },
};

</script>
