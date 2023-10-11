/* eslint-disable import/order */
import '@/@iconify/icons-bundle'
import App from '@/App.vue'
import vuetify from '@/plugins/vuetify'
import { loadFonts } from '@/plugins/webfontloader'
import router from '@/router'
import '@core/scss/template/index.scss'
import '@layouts/styles/index.scss'
import '@styles/styles.scss'
import { createPinia } from 'pinia'
import { createApp } from 'vue'
import { listen } from '@tauri-apps/api/event';

loadFonts()

// Create vue app
const app = createApp(App)

let unlisten: () => void;

async function setupGlobalListeners () {
    unlisten = await listen('open-project-event', (event: { payload: { message: string } }) => {
        console.log(event.payload.message);
    });
}

setupGlobalListeners();

// Use plugins
app.use(vuetify)
app.use(createPinia())
app.use(router)

// Mount vue app
app.mount('#app')

// Optional: Cleanup global listeners on app unmount
app.unmount = (function () {
    const cachedUnmount = app.unmount.bind(app);
    return function () {
        unlisten();
        cachedUnmount();
    }
})();