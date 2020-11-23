<template>
  <v-app>
    <v-navigation-drawer
      v-if="isLoggedIn"
      v-model="drawer"
      app
      clipped
      dark
    >
      <v-list>
        <v-list-item
          v-for="item in visibleItems"
          :key="item.title"
          :to="item.path"
          two-line
        >
          <v-list-item-action>
            <v-icon v-text="item.icon" />
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title
              :data-test="item.icon"
              v-text="item.title"
            />
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

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
      <v-menu
        transition="scale-transition"
        origin="top left"
        offset-y
      >
        <template #activator="{ on }">
          <v-chip v-on="on">
            <v-icon left>
              mdi-server
            </v-icon>
            My Device Fleet
            <v-icon right>
              mdi-chevron-down
            </v-icon>
          </v-chip>
        </template>
      </v-menu>
      <v-chip>
        <v-icon>help</v-icon>
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
            {{ $store.getters["auth/currentUser"] }}
            <v-icon right>
              mdi-chevron-down
            </v-icon>
          </v-chip>
        </template>

        <v-card>
          <v-list-item three-line>
            <v-list-item-content>
              <v-list-item-title
                class="mb-1"
                data-test="tenantID-field"
              >
                Tenant ID
              </v-list-item-title>
              <v-list-item-subtitle>
                <v-chip>
                  <span
                    data-test="tenantID-text"
                  >
                    {{ tenant }}
                  </span>
                  <v-icon
                    v-clipboard="tenant"
                    v-clipboard:success="() => {
                      this.$store.dispatch('snackbar/showSnackbarCopy', this.$copy.tenantId);
                    }"
                    right
                  >
                    mdi-content-copy
                  </v-icon>
                </v-chip>
              </v-list-item-subtitle>
            </v-list-item-content>
          </v-list-item>

          <v-divider />

          <v-list-item
            v-for="(item, index) in menu"
            :key="index"
            router
            :data-test="item.title"
            @click.prevent="triggerClick(item)"
          >
            <v-list-item-title>
              {{ item.title }}
            </v-list-item-title>
          </v-list-item>
        </v-card>
      </v-menu>
    </v-app-bar>
    <v-main class="grey lighten-4">
      <v-container
        class="pa-8"
        fluid
      >
        <router-view :key="$route.fullPath" />
      </v-container>

      <snackbar />
    </v-main>
  </v-app>
</template>

<script>

import Notification from '@/components/app_bar/notification/Notification';

export default {
  name: 'App',

  components: {
    Notification,
  },

  data() {
    return {
      drawer: true,
      clipped: false,
      items: [
        {
          icon: 'dashboard',
          title: 'Dashboard',
          path: '/',
        },
        {
          icon: 'devices',
          title: 'Devices',
          path: '/devices',
        },
        {
          icon: 'history',
          title: 'Sessions',
          path: '/sessions',
        },
        {
          icon: 'security',
          title: 'Firewall Rules',
          path: '/firewall/rules',
          hidden: !this.$env.isHosted,
        },
        {
          icon: 'mdi-account-key',
          title: 'Public Key',
          path: '/sshkeys/public_keys',
        },
      ],
      menu: [
        {
          title: 'Settings',
          type: 'path',
          path: '/settings',
          items: [{ title: 'Profile', path: '/settings/profile' }],
        },
        {
          title: 'Logout',
          type: 'method',
          method: 'logout',
        },
      ],
      admins: [
        ['Management', 'people_outline'],
        ['Settings', 'settings'],
      ],
    };
  },

  computed: {
    tenant() {
      return this.$store.getters['auth/tenant'];
    },

    isLoggedIn() {
      return this.$store.getters['auth/isLoggedIn'];
    },

    visibleItems() {
      return this.items.filter((item) => !item.hidden);
    },
  },

  methods: {
    logout() {
      this.$store.dispatch('auth/logout').then(() => {
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
  },
};

</script>

<style>
.v-list-active {
  border-left: 4px solid var(--v-primary-base);
}

.text-shadow {
  text-shadow: #000 0 0 6px;
  color: transparent;
}
</style>
